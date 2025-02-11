package models

type AuthReqest struct {
	username string
	password string
}

type AuthResponse struct {
	token string
}
