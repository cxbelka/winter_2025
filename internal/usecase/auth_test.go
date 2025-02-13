package usecase

import (
	"context"
	"crypto/sha512"
	"errors"
	"testing"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Auth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	type _tc struct {
		rq   *models.AuthReqest
		resp *models.AuthResponse
		err  error
		init func(*_tc) authRepo
	}
	testCases := map[string]_tc{
		"user_exist": {
			rq:   &models.AuthReqest{Username: "u1", Password: "p1"},
			resp: &models.AuthResponse{},
			err:  nil,
			init: func(t *_tc) authRepo {
				t.resp.Token, _ = token.Create(t.rq.Username)

				hash := sha512.New()
				hash.Write([]byte(t.rq.Password))

				mock := NewMockauthRepo(ctrl)

				mock.EXPECT().CheckLogin(ctx, t.rq.Username).Return(hash.Sum(nil), nil)
				return mock
			},
		},
		"user_not_exist": {
			rq:   &models.AuthReqest{Username: "u10", Password: "p10"},
			resp: &models.AuthResponse{},
			err:  nil,
			init: func(t *_tc) authRepo {
				t.resp.Token, _ = token.Create(t.rq.Username)

				mock := NewMockauthRepo(ctrl)

				mock.EXPECT().CheckLogin(ctx, t.rq.Username).Return(nil, models.ErrNoRows)

				mock.EXPECT().CreateUser(ctx, t.rq.Username, t.rq.Password).Return(nil)
				return mock
			},
		},
		"bad_password": {
			rq:   &models.AuthReqest{Username: "u20", Password: "NOT_p20"},
			resp: nil,
			err:  models.ErrInvalidPassword,
			init: func(t *_tc) authRepo {

				hash := sha512.New()
				hash.Write([]byte("p20"))

				mock := NewMockauthRepo(ctrl)

				mock.EXPECT().CheckLogin(ctx, t.rq.Username).Return(hash.Sum(nil), nil)
				return mock
			},
		},
		"db_issue": {
			rq:   &models.AuthReqest{Username: "u20", Password: "p20"},
			resp: nil,
			err:  errors.New("fake error"),
			init: func(t *_tc) authRepo {

				mock := NewMockauthRepo(ctrl)

				// Ожидаем что репо возвращает ошибку в обертке (models.Err.....)
				mock.EXPECT().CheckLogin(ctx, t.rq.Username).Return(nil, t.err)
				return mock
			},
		},
		"create_issue": {
			rq:   &models.AuthReqest{Username: "u20", Password: "p20"},
			resp: nil,
			err:  errors.New("fake error"),
			init: func(t *_tc) authRepo {

				mock := NewMockauthRepo(ctrl)

				mock.EXPECT().CheckLogin(ctx, t.rq.Username).Return(nil, models.ErrNoRows)

				mock.EXPECT().CreateUser(ctx, t.rq.Username, t.rq.Password).Return(t.err)
				return mock
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {

			uc := NewAuth(tc.init(&tc))

			resp, err := uc.Authorize(ctx, tc.rq)
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.resp, resp)
		})
	}
}
