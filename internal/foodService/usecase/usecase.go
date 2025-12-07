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

func (uc *serviceUC) CreateOrder(ctx context.Context, userID string) (int64, error) {
	orderId, err := uc.serviceRepo.CreateOrder(ctx, userID)
	if err != nil {
		return 0, err
	}
	err = uc.serviceRepo.PublishNewOrder(userID)
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func (uc *serviceUC) GetOrders(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error) {
	return uc.serviceRepo.GetOrdersForUser(ctx, userID, isActive)
}
