package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/cxbelka/winter_2025/internal/models"
)

type shop struct {
	db *pgx.Conn
}

func NewShop(db *pgx.Conn) *shop { //nolint:revive
	return &shop{db: db}
}

func (s *shop) BuyItem(ctx context.Context, buyer string, item string) error {
	_, err := s.db.Exec(ctx, `
	   WITH pr AS (
		INSERT INTO merch_shop.purchases (name, item, sum)
			SELECT $1, i.name, i.price
			FROM merch_shop.items AS i
			WHERE i.name = $2
		RETURNING sum
	   )
	   UPDATE merch_shop.auth SET balance = balance - (SELECT sum FROM pr) WHERE login = $1;
		`, buyer, item)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "positive_balance" {
				return errors.Join(models.ErrNoMoney, err)
			}
		}

		return errors.Join(models.ErrGeneric, err)
	}

	return nil
}

func (s *shop) ListPurchases(ctx context.Context, user string) ([]models.InventoryItem, error) {
	var purch []models.InventoryItem
	rows, err := s.db.Query(ctx, `
		SELECT item, count(purchases.item) AS qty 
		FROM merch_shop.purchases
		WHERE name = $1
		GROUP BY item
		ORDER BY qty DESC
		`, user)
	if err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	defer rows.Close()
	for rows.Next() {
		var v models.InventoryItem
		if err := rows.Scan(&v.Type, &v.Qty); err != nil {
			return nil, errors.Join(models.ErrGeneric, err)
		}
		purch = append(purch, v)
	}

	return purch, nil
}
