package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"

	"github.com/google/uuid"
)

type UseCase interface {
	CreateOrder(ctx context.Context, userID string) (uuid.UUID, error)
	GetOrders(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
	CreateSession(ctx context.Context, sessionID, userID string) error
}
