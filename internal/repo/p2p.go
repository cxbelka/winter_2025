package repo

import (
	"context"
	"errors"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type p2p struct {
	db *pgx.Conn
}

func NewP2p(db *pgx.Conn) *p2p {
	return &p2p{db: db}
}

func (p *p2p) Transfer(ctx context.Context, from string, to string, amount int) error {

	_, err := p.db.Exec(ctx, `
		WITH 
			ftx AS (UPDATE merch_shop.auth SET balance = balance-$3 WHERE login=$1),
			ttx AS (UPDATE merch_shop.auth SET balance = balance+$3 WHERE login=$2)
		INSERT INTO merch_shop.transfers (src,dst,sum) VALUES ($1, $2, $3)
		`, from, to, amount)
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

func (p *p2p) ListReceived(ctx context.Context, user string) ([]models.ReceivedTransfer, error) {
	var resive []models.ReceivedTransfer
	rows, err := p.db.Query(ctx, `
		SELECT src, sum(transfers.sum) AS sum 
		FROM merch_shop.transfers
		WHERE dst = $1
		GROUP BY src
		ORDER BY src
		`, user)
	if err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	defer rows.Close()
	for rows.Next() {
		var v models.ReceivedTransfer
		if err := rows.Scan(&v.From, &v.Amount); err != nil {
			return nil, errors.Join(models.ErrGeneric, err)
		}
		resive = append(resive, v)
	}

	return resive, rows.Err()
}

func (p *p2p) ListSent(ctx context.Context, user string) ([]models.SentTransfer, error) {
	var sent []models.SentTransfer
	rows, err := p.db.Query(ctx, `
		SELECT dst, sum(transfers.sum) AS sum 
		FROM merch_shop.transfers
		WHERE src = $1
		GROUP BY dst
		ORDER BY dst
		`, user)
	if err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	defer rows.Close()
	for rows.Next() {
		var v models.SentTransfer
		if err := rows.Scan(&v.To, &v.Amount); err != nil {
			return nil, errors.Join(models.ErrGeneric, err)
		}
		sent = append(sent, v)
	}

	return sent, nil
}
