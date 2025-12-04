package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
)

type UseCase interface {
	CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error)
	GetOrders(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)
	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
}
