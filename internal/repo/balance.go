package repo

import (
	"context"
	"errors"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/jackc/pgx/v5"
)

type balance struct {
	db *pgx.Conn
}

func NewBalance(db *pgx.Conn) *balance {
	return &balance{db: db}
}

func (b *balance) GetBalance(ctx context.Context, name string) (int, error) {
	var amount int
	err := b.db.QueryRow(ctx, `SELECT balance FROM merch_shop.auth WHERE login = $1`, name).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrNoRows
		}
		return 0, errors.Join(models.ErrGeneric, err)
	}
	return amount, nil
}
