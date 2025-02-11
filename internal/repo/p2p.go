package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type p2p struct {
	db *pgx.Conn
}

func NewP2p(db *pgx.Conn) *p2p {
	return &p2p{db: db}
}

func (p *p2p) Transfer(ctx context.Context, from string, to string, amount int) error {
	return errors.New("Unimplemented")
}
