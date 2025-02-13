package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Auth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := map[string]struct {
		rqBody string

		respCode int
		respBody string

		init func(*handle)
	}{
		"unmarshal_err": {
			rqBody: `....`,

			respCode: 500,
			respBody: `{"errors":"Internal server error"}`,
		},
		"valid_auth": {
			rqBody: `{"username":"u2","password":"p2"}`,

			respCode: 200,
			respBody: `{"token":"abc"}`,

			init: func(h *handle) {
				mock := NewMockauthUsecase(ctrl)

				mock.EXPECT().Authorize(gomock.Any(), &models.AuthReqest{Username: "u2", Password: "p2"}).Return(&models.AuthResponse{Token: "abc"}, nil)

				h.auth = mock
			},
		},
		"invalid_pass_auth": {
			rqBody: `{"username":"u2","password":"p10"}`,

			respCode: 401,
			respBody: `{"errors":"Unauthorized"}`,

			init: func(h *handle) {
				mock := NewMockauthUsecase(ctrl)

				mock.EXPECT().Authorize(gomock.Any(), &models.AuthReqest{Username: "u2", Password: "p10"}).Return(nil, models.ErrInvalidPassword)

				h.auth = mock
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			h := &handle{}

			if tc.init != nil {
				tc.init(h)
			}

			resp := httptest.NewRecorder()
			rq, err := http.NewRequest(http.MethodPost, `/api/auth`, bytes.NewBufferString(tc.rqBody))
			require.NoError(t, err)

			h.handleAuth(resp, rq)

			require.Equal(t, tc.respBody, strings.Trim(resp.Body.String(), "\n"))
			require.Equal(t, tc.respCode, resp.Code)
		})
	}
}

func Test_Transfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type _tc struct {
		rqBody   string
		userName string

		respCode int
		respBody string

		init func(*handle, *_tc)
	}

	testCases := map[string]_tc{
		"unmarshal_err": {
			rqBody: `....`,

			respCode: 500,
			respBody: `{"errors":"Internal server error"}`,
		},
		"valid_transfer": {
			rqBody:   `{"toUser":"u1","amount":30}`,
			userName: "u2",
			respCode: 200,
			respBody: ``,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Transfer(gomock.Any(), tc.userName, "u1", 30).Return(nil)

				h.acc = mock
			},
		},
		"invalid_transfer": {
			rqBody:   `{"toUser":"u1","amount":30}`,
			userName: "u2",
			respCode: 500,
			respBody: `{"errors":"Internal server error"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Transfer(gomock.Any(), tc.userName, "u1", 30).Return(models.ErrGeneric)

				h.acc = mock
			},
		},
		"not_enough_money": {
			rqBody:   `{"toUser":"u1","amount":30}`,
			userName: "u2",
			respCode: 400,
			respBody: `{"errors":"Not enough coins"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Transfer(gomock.Any(), tc.userName, "u1", 30).Return(models.ErrNoMoney)

				h.acc = mock
			},
		},
		"invalid_toUser": {
			rqBody:   `{"toUser":"u50","amount":30}`,
			userName: "u2",
			respCode: 400,
			respBody: `{"errors":"Bad request"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Transfer(gomock.Any(), tc.userName, "u50", 30).Return(models.ErrNoRows)

				h.acc = mock
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			h := &handle{}

			if tc.init != nil {
				tc.init(h, &tc)
			}

			resp := httptest.NewRecorder()
			rq, err := http.NewRequest(http.MethodPost, `/api/sendCoin`, bytes.NewBufferString(tc.rqBody))
			require.NoError(t, err)

			h.handleTransfer(resp, rq.WithContext(token.ContextWithUser(rq.Context(), tc.userName)))

			require.Equal(t, tc.respBody, strings.Trim(resp.Body.String(), "\n"))
			require.Equal(t, tc.respCode, resp.Code)
		})
	}
}

func Test_Buy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type _tc struct {
		userName string
		item     string

		respCode int
		respBody string

		init func(*handle, *_tc)
	}

	testCases := map[string]_tc{
		"valid_buy": {
			userName: "u2",
			item:     "hoody",
			respCode: 200,
			respBody: ``,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Buy(gomock.Any(), tc.userName, tc.item).Return(nil)

				h.acc = mock
			},
		},
		"invalid_buy": {
			userName: "u2",
			item:     "hoody",
			respCode: 500,
			respBody: `{"errors":"Internal server error"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Buy(gomock.Any(), tc.userName, tc.item).Return(models.ErrGeneric)

				h.acc = mock
			},
		},
		"not_enough_money": {
			userName: "u2",
			item:     "hoody",
			respCode: 400,
			respBody: `{"errors":"Not enough coins"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Buy(gomock.Any(), tc.userName, tc.item).Return(models.ErrNoMoney)

				h.acc = mock
			},
		},
		"invalid_items": {
			item:     "....",
			userName: "u2",
			respCode: 400,
			respBody: `{"errors":"Bad request"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Buy(gomock.Any(), tc.userName, tc.item).Return(models.ErrNoRows)

				h.acc = mock
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			h := &handle{}

			if tc.init != nil {
				tc.init(h, &tc)
			}

			resp := httptest.NewRecorder()
			rq, err := http.NewRequest(http.MethodGet, `/api/buy/`, nil)
			require.NoError(t, err)

			rq.SetPathValue("item", tc.item)

			h.handleBuy(resp, rq.WithContext(token.ContextWithUser(rq.Context(), tc.userName)))

			require.Equal(t, tc.respBody, strings.Trim(resp.Body.String(), "\n"))
			require.Equal(t, tc.respCode, resp.Code)
		})
	}
}

func Test_Info(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type _tc struct {
		userName string

		respCode int
		respBody string

		init func(*handle, *_tc)
	}

	testCases := map[string]_tc{

		"valid_info1": {

			userName: "u2",
			respCode: 200,
			respBody: `{"coins":400,"inventory":null,"coinHistory":{"received":null,"sent":null}}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Info(gomock.Any(), tc.userName).Return(&models.InfoResponse{Balance: 400}, nil)

				h.acc = mock
			},
		},
		"invalid_info": {

			userName: "u2",
			respCode: 500,
			respBody: `{"errors":"Internal server error"}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Info(gomock.Any(), tc.userName).Return(nil, models.ErrGeneric)

				h.acc = mock
			},
		},
		"valid_balance": {

			userName: "u2",
			respCode: 200,
			respBody: `{"coins":400,"inventory":null,"coinHistory":{"received":null,"sent":null}}`,

			init: func(h *handle, tc *_tc) {
				mock := NewMockaccountantUsecase(ctrl)

				mock.EXPECT().Info(gomock.Any(), tc.userName).Return(&models.InfoResponse{Balance: 400, Inventory: nil, Transfers: models.InfoResponseTransfers{Received: nil, Sent: nil}}, nil)

				h.acc = mock
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			h := &handle{}

			if tc.init != nil {
				tc.init(h, &tc)
			}

			resp := httptest.NewRecorder()
			rq, err := http.NewRequest(http.MethodGet, `/api/info`, nil)
			require.NoError(t, err)

			rq = rq.WithContext(token.ContextWithUser(rq.Context(), tc.userName))
			h.handleInfo(resp, rq)

			require.Equal(t, tc.respBody, strings.Trim(resp.Body.String(), "\n"))
			require.Equal(t, tc.respCode, resp.Code)
		})
	}
}

/*
HTTP/1.1 200 OK
Date: Thu, 13 Feb 2025 16:35:46 GMT
Content-Length: 102
Content-Type: text/plain; charset=utf-8
Connection: close

{
  "coins": 400,
  "inventory": [
    {
      "type": "hoody",
      "quantity": 2
    }
  ],
  "coinHistory": {
    "received": null,
    "sent": null
  }
}
  ---------------
{
  "coins": 360,
  "inventory": [
    {
      "type": "hoody",
      "quantity": 2
    }
  ],
  "coinHistory": {
    "received": null,
    "sent": [
      {
        "toUser": "u1",
        "amount": 40
      }
    ]
  }
}
*/
