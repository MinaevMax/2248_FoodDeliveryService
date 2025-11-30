package usecase

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"context"
	"log/slog"
)

type serviceUC struct {
	serviceRepo foodservice.Repository
	log         *slog.Logger
}

func NewServiceUC(serviceRepo foodservice.Repository, log *slog.Logger) foodservice.UseCase {
	return &serviceUC{
		serviceRepo: serviceRepo,
		log:         log,
	}
}

func (uc *serviceUC) CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error) {
	return uc.serviceRepo.CreateOrder(ctx, orderData)
}

func (uc *serviceUC) GetOrders(ctx context.Context, userID string) ([]*models.OrderInfo, error) {
	return uc.serviceRepo.GetOrdersForUser(ctx, userID)
}
