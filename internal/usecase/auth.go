package usecase

import (
	"context"
	"errors"

	"github.com/cxbelka/winter_2025/internal/models"
)

type auth struct {
	repo authRepo
}

type authRepo interface {
	CheckLogin(ctx context.Context, login string) (string, error)
	CreateUser(ctx context.Context, login string, pass string) error
}

func NewAuth(repo authRepo) *auth {
	return &auth{repo: repo}
}

func (a *auth) Authorize(ctx context.Context, rq *models.AuthReqest) (resp *models.AuthResponse, err error) {
	return resp, errors.New("Unimplemented")
}
