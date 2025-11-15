package usecase

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"log/slog"
)

type serviceUC struct {
	serviceRepo   foodservice.Repository
	log         *slog.Logger
}

func NewServiceUC(serviceRepo foodservice.Repository, log *slog.Logger) foodservice.UseCase {
	return &serviceUC{
		serviceRepo:   serviceRepo,
		log:         log,
	}
}