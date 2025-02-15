package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/logger"
	"github.com/cxbelka/winter_2025/internal/models"
	"github.com/cxbelka/winter_2025/internal/token"
)

func (h *handle) handleBuy(w http.ResponseWriter, r *http.Request) {
	item := r.PathValue("item")
	user := token.UserFromContext(r.Context())
	logger.AddField(r.Context(), "item", item)

	if err := h.acc.Buy(r.Context(), user, item); err != nil {
		handleError(r.Context(), w, err)
	}
}

func (h *handle) handleTransfer(w http.ResponseWriter, r *http.Request) {
	from := token.UserFromContext(r.Context())
	rq := &models.SentTransfer{}
	if err := json.NewDecoder(r.Body).Decode(rq); err != nil {
		handleError(r.Context(), w, err)

		return
	}
	logger.AddField(r.Context(), "to", rq.To)
	if err := h.acc.Transfer(r.Context(), from, rq.To, rq.Amount); err != nil {
		handleError(r.Context(), w, err)
	}
}

func (h *handle) handleInfo(w http.ResponseWriter, r *http.Request) {
	user := token.UserFromContext(r.Context())
	info, err := h.acc.Info(r.Context(), user)
	if err != nil {
		handleError(r.Context(), w, err)

		return
	}
	if err := json.NewEncoder(w).Encode(info); err != nil {
		handleError(r.Context(), w, err)
	}
}
