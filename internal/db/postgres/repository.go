package postgres

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type ServiceRepo struct {
	db *sqlx.DB
}

func NewServiceRepo(db *sqlx.DB) *ServiceRepo {
	return &ServiceRepo{db: db}
}

// Проверка, существует ли уже пользователь
func (r *ServiceRepo) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	user := &models.UserData{}
	query := `SELECT id, login, password, created_at FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, user, query, login)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Регистрация нового пользователя
func (r *ServiceRepo) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Добавляем пользователя в базу данных
	query := `INSERT INTO users (login, password, created_at) VALUES ($1, $2, NOW()) RETURNING id`
	var userID string
	err = r.db.GetContext(ctx, &userID, query, login, hashedPassword)
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

func (r *ServiceRepo) GetSessionByUserID(ctx context.Context, userID string) (*models.UserData, error) {
	session := &models.UserData{}
	query := `SELECT session_id, user_id, expires_at FROM sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	if err := r.db.GetContext(ctx, session, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // сессии нет
		}
		return nil, err
	}
	return session, nil
}

func (r *ServiceRepo) UpdateSessionExpiry(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET expires_at = NOW() + INTERVAL '15 minutes' WHERE session_id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}
