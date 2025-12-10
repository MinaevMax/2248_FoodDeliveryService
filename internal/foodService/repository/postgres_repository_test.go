package repository

import (
	"2248_FoodDeliveryService/internal/models"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO добавить testsetup

func TestServiceRepo_CreateOrder(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("CreateOrder - Success", func(t *testing.T) {
		userID := uuid.New()
		orderID := uuid.New()

		mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(orderID))

		createdOrderID, err := serviceRepo.CreateOrder(context.Background(), userID.String())

		require.NoError(t, err)
		require.Equal(t, orderID, createdOrderID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateOrder - Error", func(t *testing.T) {
		userID := uuid.New()

		mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("abc123"))

		createdOrderID, err := serviceRepo.CreateOrder(context.Background(), userID.String())

		require.Error(t, err)
		require.Equal(t, uuid.UUID{}, createdOrderID)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_ChangeOrderStatus(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("ChangeOrderStatus - Success", func(t *testing.T) {
		orderID := uuid.New()
		status := "ARRIVING"

		order := &models.ChangeOrderStatusData{
			OrderID:   orderID.String(),
			NewStatus: status,
		}

		mock.
			//ExpectPrepare(changeOrderStatusQuery).
			ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
			WithArgs(order.NewStatus, order.OrderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := serviceRepo.ChangeOrderStatus(context.Background(), order)

		require.NoError(t, err)
	})
}

func TestServiceRepo_GetOrdersForUser(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	orderID := uuid.New()
	userID := uuid.New()
	status := "ARRIVING"
	updatedAt := time.Now()

	t.Run("GetOrdersForUser - isActive", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		mock.
			ExpectQuery(getActiveOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := serviceRepo.GetOrdersForUser(context.Background(), userID.String(), true)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})

	t.Run("GetOrdersForUser - inActive", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		mock.
			ExpectQuery(getOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := serviceRepo.GetOrdersForUser(context.Background(), userID.String(), false)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})
}

func TestServiceRepo_CreateUser(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("CreateUser - Success", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now()
		user := &models.UserData{
			Login:    "user",
			Password: "pass",
			Email:    "a@b.v",
			Phone:    "+1234567890",
			IsActive: true,
		}

		mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow("user-123", user.Login, user.Password, user.Email, user.Phone, user.IsActive, now, now))

		result, err := serviceRepo.CreateUser(ctx, user)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user-123", result.ID)
		assert.Equal(t, user.Login, result.Login)
		assert.Equal(t, user.Email, result.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateUser - Database Error", func(t *testing.T) {
		ctx := context.Background()
		user := &models.UserData{
			Login:    "user2",
			Password: "pass2",
			Email:    "d@e.f",
			Phone:    "+1234567891",
			IsActive: true,
		}

		mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnError(errors.New("database error"))

		result, err := serviceRepo.CreateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateUser - Scan Error", func(t *testing.T) {
		ctx := context.Background()
		user := &models.UserData{
			Login:    "user3",
			Password: "pass3",
			Email:    "g@h.j",
			Phone:    "+1234567892",
			IsActive: true,
		}

		mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow("user-123", user.Login, user.Password, user.Email, user.Phone, "not-a-bool", time.Now(), time.Now()))

		result, err := serviceRepo.CreateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_UpdateUser(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("UpdateUser - Success", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now()
		user := &models.UserData{
			ID:       "user1",
			Login:    "newlogin",
			Email:    "a@b.c",
			Phone:    "+1234567890",
			IsActive: false,
		}

		mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, "hashedpassword", user.Email, user.Phone, user.IsActive, now.Add(-24*time.Hour), now))

		result, err := serviceRepo.UpdateUser(ctx, user)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, user.Login, result.Login)
		assert.Equal(t, user.Email, result.Email)
		assert.Equal(t, user.IsActive, result.IsActive)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateUser - User Not Found", func(t *testing.T) {
		ctx := context.Background()
		user := &models.UserData{
			ID:       "user99",
			Login:    "newlogin",
			Email:    "a@b.c",
			Phone:    "+1234567890",
			IsActive: false,
		}
		mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := serviceRepo.UpdateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateUser - Database Error", func(t *testing.T) {
		ctx := context.Background()
		user := &models.UserData{
			ID:       "user1",
			Login:    "newlogin",
			Email:    "a@b.c",
			Phone:    "+1234567890",
			IsActive: false,
		}

		mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnError(errors.New("database error"))

		result, err := serviceRepo.UpdateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_GetUserByID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("GetUserByID - Success", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New().String()
		now := time.Now()

		mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(userID, "user", "pass", "a@b.c", "+1234567890", true, now.Add(-24*time.Hour), now))

		user, err := serviceRepo.GetUserByID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "user", user.Login)
		assert.Equal(t, "a@b.c", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserByID - Not Found", func(t *testing.T) {
		ctx := context.Background()
		userID := "non-existent"

		mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := serviceRepo.GetUserByID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, user)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserByID - Database Error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		user, err := serviceRepo.GetUserByID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, user)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_RegisterUser(t *testing.T) {} //todo

func TestServiceRepo_GetUserByLogin(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("GetUserByLogin - Success", func(t *testing.T) {
		ctx := context.Background()
		login := "testuser"
		now := time.Now()

		mock.ExpectQuery(getUserByLoginQuery).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
				AddRow("user-123", login, "hashedpassword", now))

		user, err := serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, login, user.Login)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserByLogin - Not Found", func(t *testing.T) {
		ctx := context.Background()
		login := "non-existent"

		mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		user, err := serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserByLogin - Database Error", func(t *testing.T) {
		ctx := context.Background()
		login := "testuser"

		mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(errors.New("database error"))

		user, err := serviceRepo.GetUserByLogin(ctx, login)

		require.Error(t, err)
		require.Nil(t, user)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_DeleteUser(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("DeleteUser - Success", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := serviceRepo.DeleteUser(ctx, userID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteUser - Not Found", func(t *testing.T) {
		ctx := context.Background()
		userID := "non-existent"

		mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := serviceRepo.DeleteUser(ctx, userID)

		require.Error(t, err)
		assert.Equal(t, "no user found with the provided ID", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteUser - Database Error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		err := serviceRepo.DeleteUser(ctx, userID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteUser - RowsAffected Error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := serviceRepo.DeleteUser(ctx, userID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_CreateSession(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("CreateSession - Success", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := serviceRepo.CreateSession(ctx, sessionID, userID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateSession - Database Error", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "session-123"
		userID := "user-123"

		mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnError(errors.New("database error"))

		err := serviceRepo.CreateSession(ctx, sessionID, userID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_GetSessionByUserID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("GetSessionByUserID - Success", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"session_id", "user_id", "created_at", "expires_at"}).
				AddRow("session-123", userID, now, expiresAt))

		session, err := serviceRepo.GetSessionByUserID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, "session-123", session.SessionID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, expiresAt, session.ExpiresAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetSessionByUserID - Not Found", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		session, err := serviceRepo.GetSessionByUserID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, session)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetSessionByUserID - Database Error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		session, err := serviceRepo.GetSessionByUserID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, session)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_UpdateSessionExpiry(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceRepo := NewServiceRepo(sqlxDB, nil, logger)

	t.Run("UpdateSessionExpiry - Success", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "session-123"

		mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateSessionExpiry - Session Not Found", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "non-existent"

		mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		assert.Equal(t, "no session found with the provided ID", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateSessionExpiry - Database Error", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "session-123"

		mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnError(errors.New("database error"))

		err := serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateSessionExpiry - RowsAffected Error", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "session-123"

		mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
