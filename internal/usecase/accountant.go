package usecase

import (
	"context"
	"errors"

	"github.com/cxbelka/winter_2025/internal/models"
)

type transactionManager interface {
	WithTransaction(ctx context.Context, callback func(context.Context) error) error
}

type balance interface {
	GetBalance(ctx context.Context, name string) (int, error)
	UpdateBalance(ctx context.Context, name string, amount int) error
}

type p2p interface {
	Transfer(ctx context.Context, from string, to string, amount int) error
}

type shop interface {
	BuyItem(ctx context.Context, buyer string, item string) error
}

type accountant struct {
	tx      transactionManager
	balance balance
	p2p     p2p
	shop    shop
}

func NewAccountant(
	tx transactionManager,
	balance balance,
	p2p p2p,
	shop shop,
) *accountant {
	return &accountant{tx: tx, balance: balance, p2p: p2p, shop: shop}
}

func (acc *accountant) Buy(ctx context.Context, user string, item string) error {
	return errors.New("Unimplemented")
}

func (acc *accountant) Transfer(ctx context.Context, from string, to string, amount int) error {
	return errors.New("Unimplemented")
}

func (acc *accountant) Info(ctx context.Context, user string) (*models.InfoResponse, error) {
	return nil, errors.New("Unimplemented")
}
