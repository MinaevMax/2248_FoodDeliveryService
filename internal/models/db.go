package models

import "time"

type ChangeOrderStatusData struct {
	OrderID   string    `db:"order_id" json:"orderID" validate:"required"`
	NewStatus string    `db:"status" json:"status" validate:"required,oneof=UNDEFINED PACKING ARRIVING COMPLETED CANCELED"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

type OrderInfo struct {
	ID        string    `db:"id" json:"orderID"`
	Status    string    `db:"status" json:"status"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type UserData struct {
	ID        string    `json:"id,omitempty"`
	Login     string    `json:"login" db:"login" validate:"required"`
	Password  string    `json:"password,omitempty" db:"password" validate:"required"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// (Максим) Я особо не вникал в структуру, просто вписал, чтобы не жаловался код
type Customer struct{}
