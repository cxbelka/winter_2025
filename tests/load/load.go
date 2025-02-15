package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/stat"
)

const (
	RPS        = 2000 // no more than ....
	Threads    = 150  // for auth/p2p. note: pg has only 100 connections
	UsersCount = 100_000

	host = "http://localhost:8080"
)

type user struct {
	login string // password equal to login
	Token string `json:"token"`

	rqTime time.Duration
	err    bool
}

var (
	users   []user
	errChan chan error
	rps     = time.NewTicker(time.Second / RPS)

	client = &http.Client{
		Timeout: time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			Dial:                (&net.Dialer{Timeout: time.Second}).Dial},
	}
)

func main() {
	log := "\n\n"
	errChan = make(chan error)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	wg := sync.WaitGroup{}
	users = make([]user, UsersCount)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range errChan {
			_ = err
			logger.Error("failed", slog.String("err", err.Error()))
		}
	}()

	var start time.Time

	start = time.Now()
	loadAuth() // запросы на создание юзеров
	log += "auth:\n" + doStat(time.Since(start)) + "\n"

	start = time.Now()
	loadAuth() // авторизация по существующим юзерам
	log += "re-auth:\n" + doStat(time.Since(start)) + "\n"

	start = time.Now()
	loadP2P() // p2p переводы
	log += "p2p:\n" + doStat(time.Since(start)) + "\n"

	//it's a shit with missed SLA
	start = time.Now()
	loadInfo() // отчеты
	log += "info:\n" + doStat(time.Since(start)) + "\n"

	close(errChan)
	wg.Wait()

	fmt.Println(log)
}

func doStat(ttl time.Duration) string {
	log := ""
	log += fmt.Sprintf("rqs: %d, time: %3.1fs\n", len(users), ttl.Seconds())

	var errs int
	durations := make([]float64, len(users))

	for i, u := range users {
		durations[i] = u.rqTime.Seconds()
		if u.err {
			errs++
		}
	}

	sort.Float64s(durations)
	q99 := stat.Quantile(0.99, stat.LinInterp, durations, nil)
	q95 := stat.Quantile(0.95, stat.LinInterp, durations, nil)
	q50 := stat.Quantile(0.5, stat.LinInterp, durations, nil)

	log += fmt.Sprintf("avg: %3.1frps, %4.2f ms/rq | Q[99: %.2f, 95: %.2f, 50: %.2f]ms\n", float64(len(users))/ttl.Seconds(), 1000*stat.Mean(durations, nil), 1000*q99, 1000*q95, 1000*q50)
	log += fmt.Sprintf("errors: %d, SLI: %.2f%%\n", errs, 100*float64(len(users)-errs)/float64(len(users)))

	return log
}

func loadAuth() {
	sem := make(chan struct{}, Threads)
	defer close(sem)
	wg := sync.WaitGroup{}

	wg.Add(1)
	wg.Add(len(users))
	go func() {
		defer wg.Done()
		for i := 0; i < len(users); i++ {
			if users[i].login == "" {
				users[i].login = strings.ReplaceAll(uuid.NewString(), "-", "")
			}
			users[i].err = false

			sem <- struct{}{}
			go func(i int) {
				defer func() { <-sem }()
				defer wg.Done()

				<-rps.C // rps limiter

				defer func(t time.Time) {
					users[i].rqTime = time.Since(t)
				}(time.Now())

				rq, _ := http.NewRequest(http.MethodPost, host+"/api/auth", bytes.NewBufferString(
					`{"username":"`+users[i].login+`","password":"`+users[i].login+`"}`))
				rq.Close = true

				resp, err := client.Do(rq)
				if err != nil {
					users[i].err = true
					errChan <- fmt.Errorf("auth: %d failed: %w", i, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					users[i].err = true
					errChan <- fmt.Errorf("auth: %d failed: code %d", i, resp.StatusCode)
					return
				}

				if err := json.NewDecoder(resp.Body).Decode(&users[i]); err != nil {
					users[i].err = true
					errChan <- fmt.Errorf("auth: %d failed: %w", i, err)
					return
				}
			}(i)
		}
	}()

	wg.Wait()
}

func loadP2P() {
	sem := make(chan struct{}, Threads)
	defer close(sem)
	wg := sync.WaitGroup{}

	rand.Seed(time.Now().UnixMicro())

	wg.Add(1)
	wg.Add(len(users))
	go func() {
		defer wg.Done()
		for i := range users {
			users[i].err = false

			sem <- struct{}{}
			go func(i int) {
				defer func() { <-sem }()
				defer wg.Done()

				<-rps.C // rps limiter

				defer func(t time.Time) {
					users[i].rqTime = time.Since(t)
				}(time.Now())

				to := users[rand.Intn(len(users))]
				rq, _ := http.NewRequest(http.MethodPost, host+"/api/sendCoin", bytes.NewBufferString(
					`{"toUser":"`+to.login+`","amount":1}`))
				rq.Header.Add("Authorization", "Bearer "+users[i].Token)

				resp, err := client.Do(rq)
				if err != nil {
					users[i].err = true
					errChan <- fmt.Errorf("p2p: %d failed: %w", i, err)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					users[i].err = true
					errChan <- fmt.Errorf("p2p: %d failed: code %d", i, resp.StatusCode)
					return
				}
			}(i)
		}
	}()

	wg.Wait()
}

func loadInfo() {
	sem := make(chan struct{}, Threads)
	defer close(sem)
	wg := sync.WaitGroup{}

	wg.Add(1)
	wg.Add(len(users))
	go func() {
		defer wg.Done()
		for i := range users {
			users[i].err = false

			sem <- struct{}{}
			go func(i int) {
				defer func() { <-sem }()
				defer wg.Done()

				<-rps.C // rps limiter

				defer func(t time.Time) {
					users[i].rqTime = time.Since(t)
				}(time.Now())

				rq, _ := http.NewRequest(http.MethodGet, host+"/api/info", nil)
				rq.Header.Add("Authorization", "Bearer "+users[i].Token)
				rq.Close = true

				resp, err := client.Do(rq)
				if err != nil {
					users[i].err = true
					errChan <- fmt.Errorf("info: %d failed: %w", i, err)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					users[i].err = true
					errChan <- fmt.Errorf("info: %d failed: code %d", i, resp.StatusCode)
					return
				}
			}(i)
		}
	}()

	wg.Wait()
}
