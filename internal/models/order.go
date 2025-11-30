package models

import "time"

type ChangeOrderStatusData struct {
	OrderID   string `db:"order_id" json:"orderID" validate:"required"`
	NewStatus string `db:"status" json:"status" validate:"required,oneof=UNDEFINED PACKING ARRIVING COMPLETED CANCELED"`
}

type OrderInfo struct {
	ID        string    `db:"id" json:"orderID"`
	Status    string    `db:"status" json:"status"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
