package repository

import (
	"2248_FoodDeliveryService/internal/rabbitmq"
	"2248_FoodDeliveryService/internal/worker_service"
	"2248_FoodDeliveryService/internal/worker_service/models"
	"encoding/json"
	"log/slog"

	"github.com/streadway/amqp"
)

type serviceRepo struct {
	rabbit   *rabbitmq.RabbitMQ
	orderCh  chan string
	log      *slog.Logger
}

func NewServiceRepo(rabbit *rabbitmq.RabbitMQ, orderCh chan string, log *slog.Logger) worker_service.Repository {
	return &serviceRepo{rabbit: rabbit, orderCh: orderCh, log: log}
}

func (r *serviceRepo) PublishOrderStatus(orderId string, newStatus string) error {
	err := r.rabbit.Channel.ExchangeDeclare("main_exchange", "direct", true, false, false, false, nil)
	if err != nil {
		r.log.Error("failed to declare an exchange", slog.Any("errors", err))
		return err
	}
	r.log.Info(orderId)
	r.log.Info(newStatus)
	body, err := json.Marshal(models.NewStatus{OrderID: orderId, NewStatus: newStatus})
	if err != nil {
		r.log.Error("failed to marshall error message", slog.Any("error", err))
		return err
	}

	err = r.rabbit.Channel.Publish(
		"main_exchange", "order-status-change", false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		r.log.Error("failed to publish", slog.Any("error", err))
		return err
	}

	r.log.Info("successfully send new order to processing service")
	return nil
}

func (r *serviceRepo) StartNewOrdersConsumer(queueName string) {
	q, err := r.rabbit.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		r.log.Error("failed to declare a queue", slog.Any("errors", err))
		return
	}
	msgs, err := r.rabbit.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false, false, false, nil)
	if err != nil {
		r.log.Error("failed to start consumer", slog.Any("errors", err))
		return
	}

	go func() {
		r.log.Info("Consumer started")
		for {
			select {
			case <-r.rabbit.Ctx.Done():
				r.log.Info("Consumer stopped")
				return
			case d, ok := <-msgs:
				if !ok {
					r.log.Error("rabbit channel closed")
					return
				}

				orderId := ""
				if err := json.Unmarshal(d.Body, &orderId); err != nil {
					r.log.Error("failed to parse rabbit message", slog.Any("errors", err))
					d.Nack(false, true) // nack-аем сообщение (возвращаем)
					continue
				}
				r.orderCh <- orderId
				d.Ack(false) // подтверждаем обработку
			}
		}
	}()
}
