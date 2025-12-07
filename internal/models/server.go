package models

type NewOrderData struct {
	UserID    string    `db:"customer_id" json:"-" validate:"required"`
	Amount    int       `db:"amount" json:"amount" validate:"required"`
}
