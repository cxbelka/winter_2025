package usecase

import (
	"context"
	"testing"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func Test_Buy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type _tc struct {
		user string
		item string

		err error

		init func(*_tc) shop
	}

	testCases := map[string]_tc{
		"success_buy": {
			user: "u1",
			item: "hoody",

			err: nil,

			init: func(t *_tc) shop {
				mock := NewMockshop(ctrl)

				mock.EXPECT().BuyItem(ctx, t.user, t.item).Return(nil)

				return mock
			},
		},
		"item_invalid": {
			user: "u1",
			item: "wrong",

			err: models.ErrGeneric,

			init: func(t *_tc) shop {
				mock := NewMockshop(ctrl)

				mock.EXPECT().BuyItem(ctx, t.user, t.item).Return(models.ErrNoRows)

				return mock
			},
		},
		"not_enough_money": {
			user: "u1",
			item: "hoody",

			err: models.ErrGeneric,

			init: func(t *_tc) shop {
				mock := NewMockshop(ctrl)

				mock.EXPECT().BuyItem(ctx, t.user, t.item).Return(models.ErrNoMoney)

				return mock
			},
		},
		"db_issue": {
			user: "u1",
			item: "hoody",

			err: models.ErrGeneric,

			init: func(t *_tc) shop {
				mock := NewMockshop(ctrl)

				mock.EXPECT().BuyItem(ctx, t.user, t.item).Return(models.ErrGeneric)

				return mock
			},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			tc := tc

			uc := NewAccountant(nil, nil, tc.init(&tc))

			err := uc.Buy(ctx, tc.user, tc.item)
			require.ErrorIs(t, err, tc.err)
		})
	}
}
func Test_Transfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type _tc struct {
		from   string
		to     string
		amount int

		err error

		init func(*_tc) p2p
	}

	testCases := map[string]_tc{
		"success_p2p": {
			from:   "u1",
			to:     "u2",
			amount: 20,
			err:    nil,

			init: func(t *_tc) p2p {
				mock := NewMockp2p(ctrl)

				mock.EXPECT().Transfer(ctx, t.from, t.to, t.amount).Return(nil)

				return mock
			},
		},
		"toUser_invalid": {
			from:   "u1",
			to:     "u2",
			amount: 20,
			err:    models.ErrGeneric,

			init: func(t *_tc) p2p {
				mock := NewMockp2p(ctrl)

				mock.EXPECT().Transfer(ctx, t.from, t.to, t.amount).Return(models.ErrNoRows)

				return mock
			},
		},
		"not_enough_money": {
			from:   "u1",
			to:     "u2",
			amount: 20,
			err:    models.ErrGeneric,

			init: func(t *_tc) p2p {
				mock := NewMockp2p(ctrl)

				mock.EXPECT().Transfer(ctx, t.from, t.to, t.amount).Return(models.ErrNoMoney)

				return mock
			},
		},
		"db_issue": {
			from:   "u1",
			to:     "u2",
			amount: 20,
			err:    models.ErrGeneric,

			init: func(t *_tc) p2p {
				mock := NewMockp2p(ctrl)

				mock.EXPECT().Transfer(ctx, t.from, t.to, t.amount).Return(models.ErrGeneric)

				return mock
			},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			tc := tc

			uc := NewAccountant(nil, tc.init(&tc), nil)

			err := uc.Transfer(ctx, tc.from, tc.to, tc.amount)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func Test_Info(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type _tc struct {
		user string
		resp *models.InfoResponse
		err  error

		init func(*_tc) *accountant
	}

	testCases := map[string]_tc{
		"success_info": {
			user: "u1",
			resp: &models.InfoResponse{Balance: 20,
				Inventory: []models.InventoryItem{{Type: "hoody", Qty: 12}},
				Transfers: models.InfoResponseTransfers{
					Received: []models.ReceivedTransfer{{From: "u2", Amount: 20}},
					Sent:     nil,
				},
			},
			err: nil,

			init: func(t *_tc) *accountant {
				mockBalance := NewMockbalance(ctrl)
				mockBalance.EXPECT().GetBalance(ctx, t.user).Return(20, nil)

				mockP2P := NewMockp2p(ctrl)
				mockP2P.EXPECT().ListReceived(ctx, t.user).Return([]models.ReceivedTransfer{0: {From: "u2", Amount: 20}}, nil)
				mockP2P.EXPECT().ListSent(ctx, t.user).Return(nil, nil)

				mockShop := NewMockshop(ctrl)
				mockShop.EXPECT().ListPurchases(ctx, t.user).Return([]models.InventoryItem{0: {Type: "hoody", Qty: 12}}, nil)

				ret := NewAccountant(mockBalance, mockP2P, mockShop)
				return ret
			},
		},
		"Balance_Error": {
			user: "u1",
			resp: nil,
			err:  models.ErrGeneric,

			init: func(t *_tc) *accountant {
				mockBalance := NewMockbalance(ctrl)
				mockBalance.EXPECT().GetBalance(ctx, t.user).Return(0, models.ErrGeneric)

				ret := NewAccountant(mockBalance, nil, nil)
				return ret
			},
		},
		"List_Purchases_Error": {
			user: "u1",
			resp: nil,
			err:  models.ErrGeneric,

			init: func(t *_tc) *accountant {
				mockBalance := NewMockbalance(ctrl)
				mockBalance.EXPECT().GetBalance(ctx, t.user).Return(20, nil)

				mockShop := NewMockshop(ctrl)
				mockShop.EXPECT().ListPurchases(ctx, t.user).Return(nil, models.ErrGeneric)

				ret := NewAccountant(mockBalance, nil, mockShop)
				return ret
			},
		},
		"List_Recive_Error": {
			user: "u1",
			resp: nil,
			err:  models.ErrGeneric,

			init: func(t *_tc) *accountant {
				mockBalance := NewMockbalance(ctrl)
				mockBalance.EXPECT().GetBalance(ctx, t.user).Return(20, nil)

				mockP2P := NewMockp2p(ctrl)
				mockP2P.EXPECT().ListReceived(ctx, t.user).Return(nil, models.ErrGeneric)

				mockShop := NewMockshop(ctrl)
				mockShop.EXPECT().ListPurchases(ctx, t.user).Return([]models.InventoryItem{0: {Type: "hoody", Qty: 12}}, nil)

				ret := NewAccountant(mockBalance, mockP2P, mockShop)
				return ret
			},
		},
		"List_Sent_Error": {
			user: "u1",
			resp: nil,
			err:  models.ErrGeneric,

			init: func(t *_tc) *accountant {
				mockBalance := NewMockbalance(ctrl)
				mockBalance.EXPECT().GetBalance(ctx, t.user).Return(20, nil)

				mockP2P := NewMockp2p(ctrl)
				mockP2P.EXPECT().ListReceived(ctx, t.user).Return([]models.ReceivedTransfer{0: {From: "u2", Amount: 20}}, nil)
				mockP2P.EXPECT().ListSent(ctx, t.user).Return(nil, models.ErrGeneric)

				mockShop := NewMockshop(ctrl)
				mockShop.EXPECT().ListPurchases(ctx, t.user).Return([]models.InventoryItem{0: {Type: "hoody", Qty: 12}}, nil)

				ret := NewAccountant(mockBalance, mockP2P, mockShop)
				return ret
			},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			tc := tc

			uc := tc.init(&tc)

			resp, err := uc.Info(ctx, tc.user)
			require.ErrorIs(t, err, tc.err)
			require.Equal(t, tc.resp, resp)
		})
	}
}
