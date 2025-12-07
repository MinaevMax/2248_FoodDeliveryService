package models

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	} `json:"user"`
}

type OrderRequest struct {
	Amount int `json:"amount"`
}
