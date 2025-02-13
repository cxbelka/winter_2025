package usecase

import (
	"context"
	"crypto/sha512"
	"errors"
	"slices"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
)

type auth struct {
	repo authRepo
}

type authRepo interface {
	CheckLogin(ctx context.Context, login string) ([]byte, error)
	CreateUser(ctx context.Context, login string, pass string) error
}

func NewAuth(repo authRepo) *auth {
	return &auth{repo: repo}
}

func (a *auth) Authorize(ctx context.Context, rq *models.AuthReqest) (resp *models.AuthResponse, err error) {
	passHash, err := a.repo.CheckLogin(ctx, rq.Username)
	if err != nil && !errors.Is(err, models.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, models.ErrNoRows) {
		if err := a.repo.CreateUser(ctx, rq.Username, rq.Password); err != nil {
			return nil, err
		}
	} else {
		hash := sha512.New()
		hash.Write([]byte(rq.Password))

		if !slices.Equal(passHash, hash.Sum(nil)) {
			return nil, models.ErrInvalidPassword
		}
	}
	resp = &models.AuthResponse{}
	resp.Token, err = token.Create(rq.Username)
	if err != nil {
		err = errors.Join(err, models.ErrGeneric)
	}

	return resp, err
}
