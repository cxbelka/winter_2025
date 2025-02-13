package handlers

import (
	"net/http"
	"strings"

	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
)

func (h *handle) authMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// достать header, раскодировать, проверить
		tokn := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if name, err := token.Check(tokn); err != nil {
			handleError(w, models.ErrInvalidPassword)
			return
		} else {
			r = r.WithContext(token.ContextWithUser(r.Context(), name))
		}

		f(w, r)
	}
}
