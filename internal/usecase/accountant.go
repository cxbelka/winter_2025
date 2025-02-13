package usecase

import (
	"context"
	"errors"

	"github.com/cxbelka/winter_2025/internal/models"
)

type balance interface {
	GetBalance(ctx context.Context, name string) (int, error)
}

type p2p interface {
	Transfer(ctx context.Context, from string, to string, amount int) error
	ListReceived(ctx context.Context, user string) ([]models.ReceivedTransfer, error)
	ListSent(ctx context.Context, user string) ([]models.SentTransfer, error)
}

type shop interface {
	BuyItem(ctx context.Context, buyer string, item string) error
	ListPurchases(ctx context.Context, user string) ([]models.InventoryItem, error)
}

type accountant struct {
	balance balance
	p2p     p2p
	shop    shop
}

func NewAccountant(
	balance balance,
	p2p p2p,
	shop shop,
) *accountant {
	return &accountant{balance: balance, p2p: p2p, shop: shop}
}

func (acc *accountant) Buy(ctx context.Context, user string, item string) error {
	if err := acc.shop.BuyItem(ctx, user, item); err != nil {
		return errors.Join(models.ErrGeneric, err)
	}

	return nil
}

func (acc *accountant) Transfer(ctx context.Context, from string, to string, amount int) error {
	if err := acc.p2p.Transfer(ctx, from, to, amount); err != nil {
		return errors.Join(models.ErrGeneric, err)
	}
	return nil
}

func (acc *accountant) Info(ctx context.Context, user string) (info *models.InfoResponse, err error) {
	info = &models.InfoResponse{}
	if info.Balance, err = acc.balance.GetBalance(ctx, user); err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	if info.Inventory, err = acc.shop.ListPurchases(ctx, user); err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	if info.Transfers.Received, err = acc.p2p.ListReceived(ctx, user); err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	if info.Transfers.Sent, err = acc.p2p.ListSent(ctx, user); err != nil {
		return nil, errors.Join(models.ErrGeneric, err)
	}
	return info, nil
}
