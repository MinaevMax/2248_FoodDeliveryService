package rabbitmq

import (
	"2248_FoodDeliveryService/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
	Ctx     context.Context
	PubConf publisherCfg

	backoffPolicy backoff.BackOff
}

type publisherCfg struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table

	QueueName string
}

const ctxTimeout = 5 * time.Second

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

	rabbitCtx, rabbitCancel := context.WithTimeout(context.Background(), ctxTimeout)

	go func() {
		<-ctx.Done()
		rabbitCancel()
		time.Sleep(ctxTimeout)
		Close(conn, ch, log)
	}()

	return &RabbitMQ{
		conn:          conn,
		Channel:       ch,
		Ctx:           rabbitCtx,
		backoffPolicy: b,
		PubConf: publisherCfg{"main_exchange", "direct", true, false, false, false, nil, "new-orders"}, // важно, чтобы совпадали поля с аналогичными у консьюмера 
	}, nil
}

func Close(conn *amqp.Connection, ch *amqp.Channel, log *slog.Logger) {
	err := ch.Close()
	if err != nil {
		log.Warn("channel closing error", slog.Any("error", err))
	}
	err = conn.Close()
	if err != nil {
		log.Warn("connection closing error", slog.Any("error", err))
	}
}
