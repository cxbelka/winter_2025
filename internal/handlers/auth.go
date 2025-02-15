package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/logger"
	"github.com/cxbelka/winter_2025/internal/models"
)

func (h *handle) handleAuth(w http.ResponseWriter, r *http.Request) {
	rq := &models.AuthReqest{}
	if err := json.NewDecoder(r.Body).Decode(rq); err != nil {
		handleError(r.Context(), w, err)

		return
	}
	err := h.validate.Struct(rq)
	if err != nil {
		handleError(r.Context(), w, err)

		return
	}
	logger.AddField(r.Context(), "login", rq.Username)

	resp, err := h.auth.Authorize(r.Context(), rq)
	if err != nil {
		handleError(r.Context(), w, err)

		return
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		logger.AddError(r.Context(), err)
	}
}
