package worker_service

type Repository interface {
	PublishOrderStatus(orderID string, status string) error
	StartNewOrdersConsumer(queueName string)
}
