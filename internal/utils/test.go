package utils

import (
	"2248_FoodDeliveryService/internal/config"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"log/slog"
	"os"
)

func SetupRabbit() (*rabbitmq.RabbitMQ, *slog.Logger) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := &config.RabbitMQConfig{"guest", "guest", "localhost", 5672, ""}
	rabbit, _ := rabbitmq.NewRabbitMQ(cfg, ctx, logger)
	rabbit.Channel.ExchangeDeclare("main_exchange", "direct", true, false, false, false, nil)

	rabbit.Channel.QueueDeclare("new-orders", true, false, false, false, nil)
	rabbit.Channel.QueueBind("new-orders", "new-orders", "main_exchange", false, nil)
	rabbit.Channel.QueuePurge("new-orders", false)

	rabbit.Channel.QueueDeclare("order-status-change", true, false, false, false, nil)
	rabbit.Channel.QueueBind("order-status-change", "order-status-change", "main_exchange", false, nil)
	rabbit.Channel.QueuePurge("order-status-change", false)

	return rabbit, logger
}
