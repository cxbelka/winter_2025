package handlers

import (
	"context"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/models"
)

type handle struct {
	auth authUsecase
	//shop
	//p2p
}

type authUsecase interface {
	Authorize(ctx context.Context, rq *models.AuthReqest) (resp *models.AuthResponse, err error)
}

func New(auth authUsecase) *http.ServeMux {
	mx := http.NewServeMux()
	h := &handle{auth: auth}
	mx.HandleFunc("POST /api/auth", h.handleAuth)
	//mx.HandleFunc() все хендлеры из сваггера реализации которых в соседних файлах
	return mx
}
