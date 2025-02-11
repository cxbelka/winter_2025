package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type shop struct {
	db *pgx.Conn
}

func NewShop(db *pgx.Conn) *shop {
	return &shop{db: db}
}

func (s *shop) BuyItem(ctx context.Context, buyer string, item string) error {
	return errors.New("Unimplemented")
}
