package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/cxbelka/winter_2025/internal/config"
	"github.com/cxbelka/winter_2025/internal/handlers"
	"github.com/cxbelka/winter_2025/internal/repo"
	"github.com/cxbelka/winter_2025/internal/usecase"
)

type app struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
	lg         zerolog.Logger

	cfg *config.Config

	dbConn *pgxpool.Pool
	mux    *http.ServeMux
}

func New() (*app, error) { //nolint:revive
	var err error
	a := &app{wg: &sync.WaitGroup{}}

	// os.Signals listener for graceful shutdown
	a.ctx, a.cancelFunc = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	a.cfg, err = config.New()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	logLevel, err := zerolog.ParseLevel(a.cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	a.lg = zerolog.New(os.Stdout)

	// поднять подключение к БД
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s",
		a.cfg.DB.User, a.cfg.DB.Pass,
		net.JoinHostPort(a.cfg.DB.Host, strconv.Itoa(a.cfg.DB.Port)),
		a.cfg.DB.DBName)

	a.dbConn, err = pgxpool.New(context.Background(), dsn)
	//a.dbConn, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	// и пингануть для проверки
	if err = a.dbConn.Ping(context.Background()); err != nil {
		return nil, err //nolint:wrapcheck
	}

	// создать слой usecase и транспорта вложенными вызовами
	a.mux = handlers.New(
		&a.lg,
		usecase.NewAuth(repo.NewAuth(a.dbConn)),
		usecase.NewAccountant(
			repo.NewBalance(a.dbConn),
			repo.NewP2p(a.dbConn),
			repo.NewShop(a.dbConn),
		),
	)

	return a, nil
}

func (a *app) Run() error {
	defer a.cancelFunc()

	var err error
	// start all
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.HTTP.Port),
		Handler:           a.mux,
		ReadHeaderTimeout: time.Second, //nolint:mnd
	}
	errCh := make(chan error)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		<-a.ctx.Done()

		// кубер по дефолту гасит через 15 секунд после sigint
		shutdownCtx, scf := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
		defer scf()

		errCh <- srv.Shutdown(shutdownCtx)
		a.wg.Wait()

		// close all
		a.dbConn.Close()
		close(errCh)
	}()

	for e := range errCh {
		err = errors.Join(err, e)
		a.cancelFunc()
	}

	return err
}
