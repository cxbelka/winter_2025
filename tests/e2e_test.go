package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/cxbelka/winter_2025/internal/app"
	"github.com/cxbelka/winter_2025/migrations"
)

func Test_E2E(t *testing.T) {
	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	env := map[string]string{
		"DATABASE_PORT":     "15432",
		"DATABASE_USER":     "postgres",
		"DATABASE_PASSWORD": "password",
		"DATABASE_NAME":     "shop",
		"DATABASE_HOST":     "localhost",
		"SERVER_PORT":       "18080",

		"POSTGRES_PASSWORD": "password",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_DB":       "shop",
	}
	for k, v := range env {
		os.Setenv(k, v)
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:13",
			ExposedPorts: []string{env["DATABASE_PORT"] + ":5432"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
			Env:          env,
		},
		Started: true,
		Logger:  testcontainers.TestLogger(t),
	})
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, pg)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"),
	)

	time.Sleep(1 * time.Second) // hack. check WaitingFor

	dbConn, err := pgx.Connect(ctx, dsn)
	require.NoErrorf(t, err, "conn failed")
	require.NoErrorf(t, dbConn.Ping(ctx), "ping failed")
	defer dbConn.Close(ctx)

	_, err = dbConn.Exec(ctx, string(migrations.Init))
	require.NoError(t, err)

	app, err := app.New()
	require.NoError(t, err)
	go func() {
		require.NoError(t, app.Run())
	}()

	time.Sleep(1 * time.Second) // allow app to start

	httpHost := "http://localhost:" + env["SERVER_PORT"]

	mainUser := struct {
		login string
		passw string
		Token string `json:"token"`
	}{login: "main", passw: "main"}

	testsChain := []struct {
		name  string
		rq    func() *http.Request
		check func(t *testing.T, resp *http.Response)
	}{
		{
			name: "auth-new",
			rq: func() *http.Request {
				rq, _ := http.NewRequest(http.MethodPost, httpHost+"/api/auth", bytes.NewBuffer([]byte(
					`{"username":"`+mainUser.login+`","password":"`+mainUser.passw+`"}`,
				)))
				return rq
			},
			check: func(t *testing.T, resp *http.Response) {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), `{"token":`)
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var i int
				err = dbConn.QueryRow(ctx,
					`SELECT count(*) FROM merch_shop.auth WHERE login=$1 AND password=SHA512($2) AND balance=1000`,
					mainUser.login, mainUser.passw).
					Scan(&i)
				require.NoError(t, err)
				require.Equal(t, 1, i)

				require.NoError(t, json.Unmarshal(body, &mainUser)) // now mainUser has auth token
			},
		},
		{
			name: "auth-slave", // next user
			rq: func() *http.Request {
				rq, _ := http.NewRequest(http.MethodPost, httpHost+"/api/auth", bytes.NewBuffer([]byte(
					`{"username":"slave","password":"slave"}`,
				)))
				return rq
			},
			check: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "auth-slave-failed", // wrong password
			rq: func() *http.Request {
				rq, _ := http.NewRequest(http.MethodPost, httpHost+"/api/auth", bytes.NewBuffer([]byte(
					`{"username":"slave","password":"slave-wrong"}`,
				)))
				return rq
			},
			check: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "master-buy-item",
			rq: func() *http.Request {
				rq, _ := http.NewRequest(http.MethodGet, httpHost+"/api/buy/powerbank", nil)
				rq.Header.Add("Authorization", "Bearer "+mainUser.Token)

				return rq
			},
			check: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var balance int
				err = dbConn.QueryRow(ctx, `SELECT balance FROM merch_shop.auth WHERE login=$1`, mainUser.login).Scan(&balance)
				require.NoError(t, err)
				require.Equal(t, 800, balance) // starter's 1000 - 200 for powerbank
			},
		},
		{
			name: "master-p2p-transfer",
			rq: func() *http.Request {
				rq, _ := http.NewRequest(http.MethodPost, httpHost+"/api/sendCoin", bytes.NewBuffer([]byte(
					`{"toUser":"slave","amount":100}`)))
				rq.Header.Add("Authorization", "Bearer "+mainUser.Token)

				return rq
			},
			check: func(t *testing.T, resp *http.Response) {
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var balance int
				err = dbConn.QueryRow(ctx, `SELECT balance FROM merch_shop.auth WHERE login=$1`, mainUser.login).Scan(&balance)
				require.NoError(t, err)
				require.Equal(t, 700, balance) // starter's 1000 - 200 for powerbank - 100 for p2p

				err = dbConn.QueryRow(ctx, `SELECT balance FROM merch_shop.auth WHERE login=$1`, "slave").Scan(&balance)
				require.NoError(t, err)
				require.Equal(t, 1100, balance) // starter's 1000 + 100 for p2p
			},
		},
	}

	sem := make(chan struct{}, 1) // strongly disallow parallel run

	for _, ti := range testsChain {
		ti := ti
		sem <- struct{}{}
		t.Run(ti.name, func(t *testing.T) {
			defer func() { <-sem }()

			rq := ti.rq()
			resp, err := http.DefaultClient.Do(rq)
			require.NoError(t, err)
			defer resp.Body.Close()
			ti.check(t, resp)
		})
	}

}
