package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
)

type UseCase interface {
	CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error)
	GetOrders(ctx context.Context, userID string) ([]*models.OrderInfo, error)
}
