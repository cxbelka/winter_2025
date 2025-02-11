package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func NewShop(db *pgx.Conn) *shop {
	return &shop{db: db}
}

func (s *shop) BuyItem(ctx context.Context, item string) error {
	return errors.New("Unimplemented")
}
