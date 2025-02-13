package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/models"
)

func (h *handle) handleAuth(w http.ResponseWriter, r *http.Request) {
	rq := &models.AuthReqest{}
	if err := json.NewDecoder(r.Body).Decode(rq); err != nil {
		handleError(w, err)
		return
	}

	var rb any
	resp, err := h.auth.Authorize(r.Context(), rq)
	w.Header().Add("Content-type", "application/json")
	rb = resp
	if err != nil {
		handleError(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(rb); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
