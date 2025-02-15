package models

type AuthReqest struct {
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,alphanum"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
