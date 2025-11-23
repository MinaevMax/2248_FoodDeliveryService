package models

type NewOrderData struct {
	UserID string `db:"customer_id" json:"userId" validate:"required"`
	Points int    `db:"amount" json:"amount" validate:"required"`
}
