package repository

import (
	customerService "2248_FoodDeliveryService/internal/customerService"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"database/sql"
	"log"
	"log/slog"

	"github.com/google/uuid"
)

type serviceRepo struct {
	postgresql *sql.DB
	rabbit     *rabbitmq.RabbitMQ
	log        *slog.Logger
}

func (s *serviceRepo) Create(ctx context.Context, customer *models.Customer, log *log.Logger) (*models.Customer, error) {
	var o models.Customer
	if err := s.postgresql.QueryRowContext(
		ctx,
		createCustomer,
		// TODO поля
	).Scan(&o); err != nil {
		log.Panic("customerService.Create.QueryRowContext")
		return nil, err
	}

	return &o, nil
}

func (s *serviceRepo) Update(ctx context.Context, customer *models.Customer, log *log.Logger) (*models.Customer, error) {
	var o models.Customer
	if err := s.postgresql.QueryRowContext(
		ctx,
		updateCustomer,
		// TODO поля
	).Scan(&o); err != nil {
		log.Panic("customerService.Update.QueryRowContext")
		return nil, err
	}

	return &o, nil
}

func (s *serviceRepo) Delete(ctx context.Context, customerID uuid.UUID, log *log.Logger) error {
	result, err := s.postgresql.ExecContext(ctx, deleteCustomer, customerID)
	if err != nil {
		log.Panic("customerService.Delete.ExecContext")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Panic("customerService.Delete.RowsAffected")
		return err
	}
	if rowsAffected == 0 {
		log.Panic("customerService.Delete.RowsAffected")
		return sql.ErrNoRows
	}

	return nil
}

func (s serviceRepo) GetByID(customerID uuid.UUID, log *log.Logger) (*models.Customer, error) {
	c := &models.Customer{}
	row := s.postgresql.QueryRow(getCustomerById, customerID)
	err := row.Scan( /*поля*/ )
	if err != nil {
		log.Panic("customerService.GetByID.QueryRow")
		return nil, err
	}
	return с, nil
}

func NewServiceRepo(postgresql *sql.DB, rabbit *rabbitmq.RabbitMQ, log *slog.Logger) customerService.Repository {
	return &serviceRepo{postgresql: postgresql, rabbit: rabbit, log: log}
}
