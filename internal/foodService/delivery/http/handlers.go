package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"log/slog"
)

type handler struct {
	log *slog.Logger
	uc  foodservice.UseCase
}

func NewHandler(uc foodservice.UseCase, log *slog.Logger) foodservice.Handler {
	return &handler{
		log: log,
		uc:  uc,
	}
}