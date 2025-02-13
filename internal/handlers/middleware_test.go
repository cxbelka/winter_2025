package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cxbelka/winter_2025/internal/token"
	"github.com/stretchr/testify/require"
)

func Test_authMdw(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(token.UserFromContext(r.Context())))
	}

	testCases := map[string]struct {
		userName string
		respCode int
		respBody string

		sleep time.Duration
	}{
		"no_header": {
			respCode: 401,
			respBody: `{"errors":"Unauthorized"}`,
		},
		"user_test": {
			userName: "test",

			respCode: 200,
			respBody: "test",
		},
		"token_expired": {
			userName: "test",

			respCode: 401,
			respBody: `{"errors":"Unauthorized"}`,

			sleep: 2 * time.Second,
		},
	}

	t.Setenv("JWT_TTL", "1s")
	t.Setenv("JWT_SECRET", "test")
	token.Reinit()

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			h := &handle{}

			resp := httptest.NewRecorder()
			rq, err := http.NewRequest(http.MethodGet, ``, nil)
			require.NoError(t, err)
			if tc.userName != "" {
				token, err := token.Create(tc.userName)
				require.NoError(t, err)
				rq.Header.Add("Authorization", "Bearer "+token)
			}

			if tc.sleep > 0 {
				time.Sleep(tc.sleep)
			}

			f := h.authMiddleware(handlerFunc)
			f(resp, rq)

			require.Equal(t, tc.respBody, strings.Trim(resp.Body.String(), "\n"))
			require.Equal(t, tc.respCode, resp.Code)
		})
	}

}
