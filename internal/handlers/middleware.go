package handlers

import "net/http"

func (h *handle) authMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// достать header, раскодировать, проверить
		f(w, r)
	}
}
