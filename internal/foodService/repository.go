package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
)

type Repository interface {
	CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error)
	ChangeOrderStatus(ctx context.Context, newStatusData *models.ChangeOrderStatusData) error
	GetOrdersForUser(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)

	PublishNewOrder(order *models.NewOrderData) error
	StartStatusChangeConsumer(queueName string)

	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)
	CreateUser(ctx context.Context, user *models.UserData) (*models.UserData, error)
	UpdateUser(ctx context.Context, user *models.UserData) (*models.UserData, error)
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*models.UserData, error)

	CreateSession(ctx context.Context, sessionID, userID string) error
	GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error)
	UpdateSessionExpiry(ctx context.Context, sessionID string) error
}
