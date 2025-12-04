package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error)
	ChangeOrderStatus(ctx context.Context, newStatusData *models.ChangeOrderStatusData) error
	GetOrdersForUser(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)

	CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, customerID uuid.UUID) error
	GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error)

	PublishNewOrder(order *models.NewOrderData) error
	StartStatusChangeConsumer(queueName string)

	GetUserByLogin(ctx context.Context, login string) (*models.UserData, error)
	RegisterUser(ctx context.Context, login, password string) (*models.UserData, error)

	GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error)
	UpdateSessionExpiry(ctx context.Context, sessionID string) error
}
