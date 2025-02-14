package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/cxbelka/winter_2025/internal/logger"
	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
)

func (h *handle) authMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// достать header, раскодировать, проверить
		tokn := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if name, err := token.Check(tokn); err != nil {

			logger.AddError(r.Context(), err)

			handleError(w, models.ErrInvalidPassword)
			return
		} else {
			logger.AddField(r.Context(), "user", name)

			r = r.WithContext(token.ContextWithUser(r.Context(), name))
		}

		f(w, r)
	}
}

type wrapper struct {
	http.ResponseWriter
	ResponStatus int
}

func (w *wrapper) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.ResponStatus = statusCode
}

func (h *handle) loggerMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		wrap := &wrapper{ResponseWriter: w, ResponStatus: http.StatusOK}

		defer func() {
			if e := recover(); e != nil {
				// log
				handleError(w, models.ErrGeneric)
			}
		}()

		lctx := h.lg.With().
			Str("method", r.Method).
			Str("path", r.URL.String()).
			Str("id", uuid.NewString())

		ctx := lctx.Logger().WithContext(r.Context())
		f(wrap, r.WithContext(ctx))

		evt := zerolog.Ctx(ctx).Info()
		if wrap.ResponStatus != http.StatusOK {
			evt = zerolog.Ctx(ctx).Error()
		}
		evt.Int("code", wrap.ResponStatus).Send()
	}
}
