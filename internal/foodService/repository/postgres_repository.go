package repository

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
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

func (s *serviceRepo) GetOrdersForUser(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error) {
	var query string
	if isActive {
		query = getActiveOrdersQuery
	} else {
		query = getOrdersQuery
	}

	rows, err := s.postgresql.QueryxContext(
		ctx,
		query,
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

func (s *serviceRepo) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error) {
	c := &models.Customer{}
	row := s.postgresql.QueryRowContext(ctx, getCustomerByIdQuery, customerID)
	err := row.Scan( /*поля*/ )
	if err != nil {
		s.log.Error("failed to get customer by id", slog.Any("error", err))
		return nil, err
	}
	return c, nil
}

// Регистрация нового пользователя
func (r *serviceRepo) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Добавляем пользователя в базу данных
	query := `INSERT INTO users (login, password, created_at) VALUES ($1, $2, NOW()) RETURNING id`
	var userID string
	err = r.postgresql.GetContext(ctx, &userID, query, login, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	user := &models.UserData{
		ID:        userID,
		Login:     login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	return user, nil
}

// Проверка, существует ли уже пользователь
func (r *serviceRepo) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	user := &models.UserData{}
	query := `SELECT id, login, password, created_at FROM users WHERE login = $1`
	err := r.postgresql.GetContext(ctx, user, query, login)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *serviceRepo) DeleteUser(ctx context.Context, userID string) error {
	// Выполняем запрос на удаление пользователя по ID
	result, err := r.postgresql.ExecContext(ctx, deleteUserQuery, userID)
	if err != nil {
		r.log.Error("failed to delete user", slog.Any("error", err))
		return err
	}

	// Проверяем, что пользователь был удален
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("failed to get rows affected", slog.Any("error", err))
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with the provided ID")
	}

	return nil
}
