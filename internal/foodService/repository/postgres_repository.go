package repository

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type serviceRepo struct {
	postgresql *sqlx.DB
	rabbit     *rabbitmq.RabbitMQ
	log        *slog.Logger
}

func NewServiceRepo(postgresql *sqlx.DB, rabbit *rabbitmq.RabbitMQ, log *slog.Logger) foodservice.Repository {
	return &serviceRepo{postgresql: postgresql, rabbit: rabbit, log: log}
}

func (s *serviceRepo) CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error) {
	result, err := s.postgresql.NamedExecContext(
		ctx,
		createOrderQuery,
		orderData,
	)
	if err != nil {
		s.log.Error("failed to create order", slog.Any("error", err))
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		s.log.Error("failed to get new order id", slog.Any("error", err))
		return 0, err
	}

	return id, nil
}

func (s *serviceRepo) ChangeOrderStatus(ctx context.Context, newStatusData *models.ChangeOrderStatusData) error {
	result, err := s.postgresql.NamedExecContext(
		ctx,
		createOrderQuery,
		newStatusData,
	)
	if err != nil {
		s.log.Error("failed to change order status", slog.Any("error", err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get delete rows affected", slog.Any("error", err))
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no orders status changed")
	}

	return nil
}

func (s *serviceRepo) GetOrdersForUser(ctx context.Context, userID string) ([]*models.OrderInfo, error) {
	rows, err := s.postgresql.QueryxContext(
		ctx,
		getOrdersQuery,
		userID,
	)
	if err != nil {
		s.log.Error("failed to get orders", slog.Any("error", err))
		return nil, err
	}

	var orders []*models.OrderInfo

	for rows.Next() {
		var order models.OrderInfo
		if err := rows.Scan(&order); err != nil {
			return nil, fmt.Errorf("failed to scan orders for client rows: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders rows: %w", err)
	}

	return orders, nil
}



func (s *serviceRepo) CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	var o models.Customer
	if err := s.postgresql.QueryRowContext(
		ctx,
		createCustomerQuery,
		// TODO поля
	).Scan(&o); err != nil {
		s.log.Error("failed to create customer", slog.Any("error", err))
		return nil, err
	}

	return &o, nil
}

func (s *serviceRepo) UpdateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	var o models.Customer
	if err := s.postgresql.QueryRowContext(
		ctx,
		updateCustomerQuery,
		// TODO поля
	).Scan(&o); err != nil {
		s.log.Error("failed to update customer", slog.Any("error", err))
		return nil, err
	}

	return &o, nil
}

func (s *serviceRepo) DeleteCustomer(ctx context.Context, customerID uuid.UUID) error {
	result, err := s.postgresql.ExecContext(ctx, deleteCustomerQuery, customerID)
	if err != nil {
		s.log.Error("failed to delete customer", slog.Any("error", err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get delete rows affected", slog.Any("error", err))
		return err
	}
	if rowsAffected == 0 {
		s.log.Error("no customers deleted")
		return sql.ErrNoRows
	}

	return nil
}

func (s serviceRepo) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error) {
	c := &models.Customer{}
	row := s.postgresql.QueryRowContext(ctx, getCustomerByIdQuery, customerID)
	err := row.Scan( /*поля*/ )
	if err != nil {
		s.log.Error("failed to get customer by id", slog.Any("error", err))
		return nil, err
	}
	return c, nil
}
