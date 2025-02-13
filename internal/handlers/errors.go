package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cxbelka/winter_2025/internal/models"
)

type handlerError struct {
	code   int    `json:"-"`
	Status string `json:"errors"`
}

var (
	errGeneric       = handlerError{code: http.StatusInternalServerError, Status: "Internal server error"}
	errUnauthorized  = handlerError{code: http.StatusUnauthorized, Status: "Unauthorized"}
	errBadRequest    = handlerError{code: http.StatusBadRequest, Status: "Bad request"}
	errNoEnoughMoney = handlerError{code: http.StatusBadRequest, Status: "Not enough coins"}
)

func handleError(w http.ResponseWriter, err error) {
	var e handlerError
	switch {
	case errors.Is(err, models.ErrNoMoney):
		e = errNoEnoughMoney
	case errors.Is(err, models.ErrGeneric):
		e = errGeneric
	case errors.Is(err, models.ErrInvalidPassword):
		e = errUnauthorized
	case errors.Is(err, models.ErrNoRows):
		e = errBadRequest

	default:
		e = errGeneric
	}

	w.WriteHeader(e.code)
	if err = json.NewEncoder(w).Encode(e); err != nil {
		fmt.Printf("%+v\n", err)
	}

}
