package models

type NewOrderData struct {
	UserID  string   `json:"userId" validate:"required"`
	Points  []string `json:"orderPoints" validate:"required"`
	Comment string   `json:"comment"`
}
