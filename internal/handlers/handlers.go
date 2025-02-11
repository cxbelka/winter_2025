package handlers

import (
	"context"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/models"
)

type handle struct {
	auth authUsecase
	acc  accountantUsecase
}

type authUsecase interface {
	Authorize(ctx context.Context, rq *models.AuthReqest) (resp *models.AuthResponse, err error)
}
type accountantUsecase interface {
	Buy(ctx context.Context, user string, item string) error
	Transfer(ctx context.Context, from string, to string, amount int) error
	Info(ctx context.Context, user string) (*models.InfoResponse, error)
}

func New(auth authUsecase, acc accountantUsecase) *http.ServeMux {
	mx := http.NewServeMux()
	h := &handle{auth: auth, acc: acc}

	mx.HandleFunc("POST /api/auth", h.handleAuth)
	mx.HandleFunc("GET /api/info", h.authMiddleware(h.handleInfo))
	mx.HandleFunc("POST /api/sendCoin", h.authMiddleware(h.handleTransfer))
	mx.HandleFunc("GET /api/buy/{item}", h.authMiddleware(h.handleBuy)) // запрос на изменение данных лучше оформлять как POST, но ТЗ требует GET

	return mx
}
