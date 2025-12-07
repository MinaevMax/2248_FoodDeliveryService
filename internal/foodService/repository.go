package foodservice

import (
	"2248_FoodDeliveryService/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateOrder(ctx context.Context, userID string) (int64, error)
	ChangeOrderStatus(ctx context.Context, newStatusData *models.ChangeOrderStatusData) error
	GetOrdersForUser(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error)

	CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, customerID uuid.UUID) error
	GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error)

	PublishNewOrder(userID string) error
	StartStatusChangeConsumer(queueName string)
}
