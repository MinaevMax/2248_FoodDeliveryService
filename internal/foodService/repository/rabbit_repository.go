package repository

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/streadway/amqp"
)

func (r *serviceRepo) PublishNewOrder(userID string) error {
	err := r.rabbit.Channel.ExchangeDeclare(r.rabbit.PubConf.Name, r.rabbit.PubConf.Kind, r.rabbit.PubConf.Durable, r.rabbit.PubConf.AutoDelete, r.rabbit.PubConf.Internal, r.rabbit.PubConf.NoWait, r.rabbit.PubConf.Args)
	if err != nil {
		r.log.Error("failed to declare an exchange", slog.Any("errors", err))
		return err
	}

	body, err := json.Marshal(map[string]string{"userID": userID})
	if err != nil {
		r.log.Error("failed to marshall error message", slog.Any("error", err))
		return err
	}

	err = r.rabbit.Channel.Publish(
		r.rabbit.PubConf.Name, r.rabbit.PubConf.QueueName, false, false,
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

func (r *serviceRepo) StartStatusChangeConsumer(queueName string) {
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

				var newOrderStatus models.ChangeOrderStatusData
				if err := json.Unmarshal(d.Body, &newOrderStatus); err != nil {
					r.log.Error("failed to parse rabbit message", slog.Any("errors", err))
					d.Nack(false, true) // nack-аем сообщение (возвращаем)
					continue
				}
				newOrderStatus.UpdatedAt = time.Now()
				if err := r.changeStatus(&newOrderStatus); r != nil {
					r.log.Error("failed to change order status", slog.Any("errors", err))
					d.Nack(false, true) // nack-аем сообщение (возвращаем)
					continue
				}
				d.Ack(false) // подтверждаем обработку
			}
		}
	}()
}

func (r *serviceRepo) changeStatus(newOrderStatus *models.ChangeOrderStatusData) error {
	changeOrderStatusCtx, changeOrderStatusCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer changeOrderStatusCancel()
	return r.ChangeOrderStatus(changeOrderStatusCtx, newOrderStatus)
}
