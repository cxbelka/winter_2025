package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cxbelka/winter_2025/internal/config"
	"github.com/cxbelka/winter_2025/internal/handlers"
	"github.com/cxbelka/winter_2025/internal/repo"
	"github.com/cxbelka/winter_2025/internal/usecase"
)

type app struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup

	cfg *config.Config

	dbConn *pgx.Conn
	mux    *http.ServeMux
}

func New() (*app, error) {
	var err error
	a := &app{wg: &sync.WaitGroup{}}

	// os.Signals listener for graceful shutdown
	a.ctx, a.cancelFunc = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	a.cfg, err = config.New()
	if err != nil {
		return nil, err
	}
	// поднять подключение к БД
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", a.cfg.DB.User, a.cfg.DB.Pass, a.cfg.DB.Host, a.cfg.DB.Port, a.cfg.DB.DBName)
	a.dbConn, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	// и пингануть для проверки
	if err = a.dbConn.Ping(context.Background()); err != nil {
		return nil, err
	}

	// создать слой usecase и транспорта вложенными вызовами
	a.mux = handlers.New(
		usecase.NewAuth(repo.NewAuth(a.dbConn)),
	)

	return a, nil
}

func (a *app) Run() error {
	defer a.cancelFunc()

	var err error
	// start all
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.HTTP.Port),
		Handler: a.mux,
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

		shutdownCtx, scf := context.WithTimeout(context.Background(), 10*time.Second)
		defer scf()

		errCh <- srv.Shutdown(shutdownCtx)
		a.wg.Wait()

		// close all
		errCh <- a.dbConn.Close(shutdownCtx)
		close(errCh)
	}()

	for e := range errCh {
		err = errors.Join(err, e)
		a.cancelFunc()
	}

	return err
}
