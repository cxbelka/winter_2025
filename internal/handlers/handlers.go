package handlers

//go:generate mockgen -package handlers -source=handlers.go -destination=handler_mocks.go *

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/cxbelka/winter_2025/internal/models"
)

type handle struct {
	lg *zerolog.Logger

	auth     authUsecase
	acc      accountantUsecase
	validate *validator.Validate
}

type authUsecase interface {
	Authorize(ctx context.Context, rq *models.AuthReqest) (resp *models.AuthResponse, err error)
}
type accountantUsecase interface {
	Buy(ctx context.Context, user string, item string) error
	Transfer(ctx context.Context, from string, to string, amount int) error
	Info(ctx context.Context, user string) (*models.InfoResponse, error)
}

func New(lg *zerolog.Logger, auth authUsecase, acc accountantUsecase) *http.ServeMux {
	mx := http.NewServeMux()
	h := &handle{lg: lg, auth: auth, acc: acc}
	h.validate = validator.New()

	mx.HandleFunc("POST /api/auth", h.loggerMiddleware(h.handleAuth))
	mx.HandleFunc("GET /api/info", h.loggerMiddleware(h.authMiddleware(h.handleInfo)))
	mx.HandleFunc("POST /api/sendCoin", h.loggerMiddleware(h.authMiddleware(h.handleTransfer)))
	// запрос на изменение данных лучше оформлять как POST, но ТЗ требует GET.
	mx.HandleFunc("GET /api/buy/{item}", h.loggerMiddleware(h.authMiddleware(h.handleBuy)))

	return mx
}
