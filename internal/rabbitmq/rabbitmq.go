package rabbitmq

import (
	"2248_FoodDeliveryService/internal/config"
	"context"
	"fmt"
	"log/slog"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	backoffPolicy backoff.BackOff
}

// New создает подключение к RabbitMQ и открывает канал
func NewRabbitMQ(cfg *config.RabbitMQConfig, ctx context.Context, log *slog.Logger) (*RabbitMQ, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	)

	b := backoff.NewExponentialBackOff()

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		err := ch.Close()
		if err != nil {
			log.Warn("channel closing error", slog.Any("error", err))
		}
		err = conn.Close()
		if err != nil {
			log.Warn("connection closing error", slog.Any("error", err))
		}
	}()

	return &RabbitMQ{
		conn:          conn,
		channel:       ch,
		backoffPolicy: b,
	}, nil
}
