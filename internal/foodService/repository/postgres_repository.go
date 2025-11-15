package repository

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"log/slog"
)

type serviceRepo struct {
	postgresql     *interface{}
	rabbit    *rabbitmq.RabbitMQ
	log       *slog.Logger
}

func NewServiceRepo(postgresql *interface{}, rabbit *rabbitmq.RabbitMQ, log *slog.Logger) foodservice.Repository {
	return &serviceRepo{postgresql: postgresql, rabbit: rabbit, log: log}
}
