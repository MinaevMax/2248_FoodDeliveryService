package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
)

type UseCase interface {
	CreateOrder(ctx context.Context, userID string) (int64, error)
	GetOrders(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)
}
