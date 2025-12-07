package repository

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"fmt"
	"log/slog"
	"time"

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

func (s *serviceRepo) CreateOrder(ctx context.Context, userID string) (int64, error) {
	result, err := s.postgresql.ExecContext(
		ctx,
		createOrderQuery,
		userID,
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

// ChangeOrderStatus обновляет статус заказа в БД
func (s *serviceRepo) ChangeOrderStatus(ctx context.Context, newStatusData *models.ChangeOrderStatusData) error {
	result, err := s.postgresql.NamedExecContext(
		ctx,
		changeOrderStatusQuery,
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

// GetOrdersForUser получает список заказов пользователя (все или только активные)
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

// CreateUser создаёт нового пользователя со всеми параметрами
func (s *serviceRepo) CreateUser(ctx context.Context, user *models.UserData) (*models.UserData, error) {
	var u models.UserData
	err := s.postgresql.QueryRowContext(ctx, createUserQuery, user.Login, user.Password, user.Email, user.Phone, user.IsActive).Scan(
		&u.ID, &u.Login, &u.Password, &u.Email, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		s.log.Error("failed to create user", slog.Any("error", err))
		return nil, err
	}

	return &u, nil
}

// UpdateUser обновляет данные пользователя в БД
func (s *serviceRepo) UpdateUser(ctx context.Context, user *models.UserData) (*models.UserData, error) {
	var u models.UserData
	err := s.postgresql.QueryRowContext(ctx, updateUserQuery, user.Login, user.Email, user.Phone, user.IsActive, user.ID).Scan(
		&u.ID, &u.Login, &u.Password, &u.Email, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		s.log.Error("failed to update user", slog.Any("error", err))
		return nil, err
	}

	return &u, nil
}

// GetUserByID получает пользователя по ID
func (s *serviceRepo) GetUserByID(ctx context.Context, userID string) (*models.UserData, error) {
	u := &models.UserData{}
	err := s.postgresql.GetContext(ctx, u, getUserByIDQuery, userID)
	if err != nil {
		s.log.Error("failed to get user by id", slog.Any("error", err))
		return nil, err
	}
	return u, nil
}

func (r *serviceRepo) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var userID string
	err = r.postgresql.GetContext(ctx, &userID, registerUserQuery, login, hashedPassword)
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

func (r *serviceRepo) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	user := &models.UserData{}
	query := `SELECT id, login, password, created_at FROM users WHERE login = $1`
	err := r.postgresql.GetContext(ctx, user, query, login)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser удаляет пользователя по ID
func (r *serviceRepo) DeleteUser(ctx context.Context, userID string) error {
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

// CreateSession создаёт сессию пользователя (24 часа TTL)
func (r *serviceRepo) CreateSession(ctx context.Context, sessionID, userID string) error {
	_, err := r.postgresql.ExecContext(ctx, createSessionQuery, sessionID, userID)
	if err != nil {
		r.log.Error("failed to create session", slog.Any("error", err))
		return err
	}
	return nil
}

// GetSessionByUserID получает активную сессию пользователя
func (r *serviceRepo) GetSessionByUserID(ctx context.Context, userID string) (*models.Session, error) {
	session := &models.Session{}
	err := r.postgresql.GetContext(ctx, session, getSessionByUserIDQuery, userID)
	if err != nil {
		r.log.Error("failed to get session by user id", slog.Any("error", err))
		return nil, err
	}
	return session, nil
}

// UpdateSessionExpiry продлевает TTL сессии на 24 часа
func (r *serviceRepo) UpdateSessionExpiry(ctx context.Context, sessionID string) error {
	result, err := r.postgresql.ExecContext(ctx, updateSessionExpiryQuery, sessionID)
	if err != nil {
		r.log.Error("failed to update session expiry", slog.Any("error", err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("failed to get rows affected", slog.Any("error", err))
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no session found with the provided ID")
	}

	return nil
}
