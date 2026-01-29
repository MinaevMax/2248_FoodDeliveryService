package models

type NewStatus struct {
	OrderID   string `json:"order_id"`
	NewStatus string `json:"status"`
}
