package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type transaction struct {
	db *pgx.Conn
}

func NewTransactionManager(db *pgx.Conn) *transaction {
	return &transaction{db: db}
}

func (t *transaction) WithTransaction(ctx context.Context, callback func(context.Context) error) error {
	return errors.New("Unimplemented")
}
