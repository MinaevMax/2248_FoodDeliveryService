package models

import (
	"time"

	"github.com/google/uuid"
)

type ChangeOrderStatusData struct {
	OrderID     string    `json:"order_id" validate:"required"`
	UUIDOrderID uuid.UUID `db:"order_id" json:"-"`
	NewStatus   string    `db:"status" json:"status" validate:"required,oneof=UNDEFINED PACKING ARRIVING COMPLETED CANCELED"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}

type OrderInfo struct {
	ID        string    `db:"id" json:"orderID"`
	Status    string    `db:"status" json:"status"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type UserData struct {
	ID        string    `json:"id,omitempty" db:"id"`
	Login     string    `json:"login" db:"login" validate:"required"`
	Password  string    `json:"-" db:"password" validate:"required"`
	Email     string    `json:"email,omitempty" db:"email"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	IsActive  bool      `json:"is_active,omitempty" db:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type Session struct {
	SessionID string    `db:"session_id" json:"sessionID"`
	UserID    string    `db:"user_id" json:"userID"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
