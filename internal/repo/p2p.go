package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func NewP2p(db *pgx.Conn) *p2p {
	return &p2p{db: db}
}

func (p *p2p) CheckSum(ctx context.Context, name string) (int, error) {
	return 0, errors.New("Unimplemented")
}
