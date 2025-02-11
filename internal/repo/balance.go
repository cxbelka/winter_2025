package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type balance struct {
	db *pgx.Conn
}

func NewBalance(db *pgx.Conn) *balance {
	return &balance{db: db}
}

func (b *balance) GetBalance(ctx context.Context, name string) (int, error) {
	return 0, errors.New("Unimplemented")
}

func (b *balance) UpdateBalance(ctx context.Context, name string, amount int) error {
	return errors.New("Unimplemented")
}
