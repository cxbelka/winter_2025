package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func NewAuth(db *pgx.Conn) *auth {
	return &auth{db: db}
}

func (a *auth) CheckLogin(ctx context.Context, login string) (string, error) {
	return "", errors.New("Unimplemented")
}
func (a *auth) CreateUser(ctx context.Context, login string, pass string) error {
	return errors.New("Unimplemented")
}
