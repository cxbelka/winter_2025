package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cxbelka/winter_2025/internal/models"
)

type auth struct {
	db *pgxpool.Pool
}

func NewAuth(db *pgxpool.Pool) *auth { //nolint:revive
	return &auth{db: db}
}

func (a *auth) CheckLogin(ctx context.Context, login string) ([]byte, error) {
	var passwd []byte
	err := a.db.QueryRow(ctx, `SELECT password FROM merch_shop.auth WHERE login = $1`, login).Scan(&passwd)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNoRows
		}

		return nil, errors.Join(models.ErrGeneric, err)
	}

	return passwd, nil
}

func (a *auth) CreateUser(ctx context.Context, login string, pass string) error {
	_, err := a.db.Exec(ctx,
		`INSERT INTO merch_shop.auth (login, password, balance) VALUES ($1, SHA512($2), 1000)`,
		login, pass)
	if err != nil {
		return errors.Join(models.ErrGeneric, err)
	}

	return nil
}
