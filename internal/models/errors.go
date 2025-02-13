package models

import "errors"

// internal errors
var (
	ErrGeneric         = errors.New("generic error")
	ErrNoRows          = errors.New("no data")
	ErrInvalidPassword = errors.New("wrong password")
	ErrNoMoney         = errors.New("not enough coins")
)
