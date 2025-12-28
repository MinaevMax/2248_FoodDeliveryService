package repository

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TODO поменять на testSuite
type testContext struct {
	sqlxDB      *sqlx.DB
	mock        sqlmock.Sqlmock
	logger      *slog.Logger
	serviceRepo foodservice.Repository
}

func (c *testContext) setupTest() {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	c.sqlxDB = sqlx.NewDb(db, "sqlmock")
	c.mock = mock
	c.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c.serviceRepo = NewServiceRepo(c.sqlxDB, nil, c.logger)
}

func (c *testContext) teardownTest() {
	c.sqlxDB.Close()
}

func TestServiceRepo_CreateOrder(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		orderID := uuid.New()

		tc.mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(orderID))

		createdOrderID, err := tc.serviceRepo.CreateOrder(context.Background(), userID.String())

		require.NoError(t, err)
		require.Equal(t, orderID, createdOrderID)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Error", func(t *testing.T) {
		userID := uuid.New()

		tc.mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("abc123"))

		createdOrderID, err := tc.serviceRepo.CreateOrder(context.Background(), userID.String())

		require.Error(t, err)
		require.Equal(t, uuid.UUID{}, createdOrderID)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_ChangeOrderStatus(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	orderID := uuid.New()
	status := "ARRIVING"

	order := &models.ChangeOrderStatusData{
		OrderID:   orderID.String(),
		NewStatus: status,
	}

	t.Run("Success", func(t *testing.T) {
		tc.mock.
			//ExpectPrepare(changeOrderStatusQuery).
			ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
			WithArgs(order.NewStatus, order.OrderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := tc.serviceRepo.ChangeOrderStatus(context.Background(), order)

		require.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		tc.mock.
			ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
			WithArgs(order.NewStatus, orderID).
			WillReturnError(fmt.Errorf("database connection failed"))

		err := tc.serviceRepo.ChangeOrderStatus(context.Background(), order)

		require.Error(t, err)
	})
}

func TestServiceRepo_GetOrdersForUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	orderID := uuid.New()
	userID := uuid.New()
	status := "ARRIVING"
	updatedAt := time.Now()

	t.Run("Active Orders", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		tc.mock.
			ExpectQuery(getActiveOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := tc.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), true)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})

	t.Run("All Orders", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		tc.mock.
			ExpectQuery(getOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := tc.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), false)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})

	t.Run("Database connection timeout", func(t *testing.T) {
		tc.mock.
			ExpectQuery(getActiveOrdersQuery).
			WithArgs(userID.String()).
			WillReturnError(fmt.Errorf("database connection timeout"))

		orders, err := tc.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), true)

		require.Error(t, err)
		require.Nil(t, orders)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_CreateUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	ctx := context.Background()
	now := time.Now()
	userID := uuid.New()
	user := &models.UserData{
		ID:       userID.String(),
		Login:    "user",
		Password: "pass",
		Email:    "a@b.v",
		Phone:    "+1234567890",
		IsActive: true,
	}

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, user.Password, user.Email, user.Phone, user.IsActive, now, now))

		result, err := tc.serviceRepo.CreateUser(ctx, user)

		require.NoError(t, err)
		require.Equal(t, userID.String(), result.ID)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Failed to create user", func(t *testing.T) {
		tc.mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, user.Password, user.Email, user.Phone, "not-a-bool", time.Now(), time.Now()))

		result, err := tc.serviceRepo.CreateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_UpdateUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	ctx := context.Background()
	now := time.Now()
	user := &models.UserData{
		ID:       "user1",
		Login:    "newlogin",
		Email:    "a@b.c",
		Phone:    "+1234567890",
		IsActive: false,
	}

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, "hashedpassword", user.Email, user.Phone, user.IsActive, now.Add(-24*time.Hour), now))

		result, err := tc.serviceRepo.UpdateUser(ctx, user)

		require.NoError(t, err)
		require.Equal(t, user.ID, result.ID)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("User Not Found", func(t *testing.T) {
		tc.mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := tc.serviceRepo.UpdateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Contains(t, err.Error(), "no rows in result set")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_GetUserByID(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		tc.mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID.String()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(userID.String(), "user", "pass", "a@b.c", "+1234567890", true, now.Add(-24*time.Hour), now))

		user, err := tc.serviceRepo.GetUserByID(ctx, userID.String())

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, userID.String(), user.ID)
		require.Equal(t, "user", user.Login)
		require.Equal(t, "a@b.c", user.Email)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		tc.mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID.String()).
			WillReturnError(sql.ErrNoRows)

		user, err := tc.serviceRepo.GetUserByID(ctx, userID.String())

		require.Error(t, err)
		require.Nil(t, user)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_RegisterUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	userID := uuid.New()
	login := "user"
	password := "password"

	t.Run("Success", func(t *testing.T) {
		tc.mock.
			ExpectQuery(registerUserQuery).
			WithArgs(login, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID.String()))

		user, err := tc.serviceRepo.RegisterUser(context.Background(), login, password)
		require.NoError(t, err)
		require.Equal(t, userID.String(), user.ID)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Error", func(t *testing.T) {
		tc.mock.
			ExpectQuery(registerUserQuery).
			WithArgs(login, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		user, err := tc.serviceRepo.RegisterUser(context.Background(), login, password)

		require.Error(t, err)
		require.Nil(t, user)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Password too long", func(t *testing.T) {
		password := "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
		_, err := tc.serviceRepo.RegisterUser(context.Background(), login, password)
		require.Error(t, err)
	})
}

func TestServiceRepo_GetUserByLogin(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	ctx := context.Background()
	login := "user"
	userID := uuid.New()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectQuery(getUserByLoginQuery).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
				AddRow(userID.String(), login, "hashedpassword", now))

		user, err := tc.serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, userID.String(), user.ID)
		require.Equal(t, login, user.Login)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		tc.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		user, err := tc.serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.Nil(t, user)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		tc.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(errors.New("database error"))

		user, err := tc.serviceRepo.GetUserByLogin(ctx, login)

		require.Error(t, err)
		require.Nil(t, user)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_DeleteUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := tc.serviceRepo.DeleteUser(ctx, userID.String())

		require.NoError(t, err)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		tc.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := tc.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Equal(t, "no user found with the provided ID", err.Error())
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		tc.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnError(errors.New("database error"))

		err := tc.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("RowsAffected Error", func(t *testing.T) {
		tc.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := tc.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_CreateSession(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()

	ctx := context.Background()
	sessionID := "session-123"
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := tc.serviceRepo.CreateSession(ctx, sessionID, userID)

		require.NoError(t, err)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		tc.mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnError(errors.New("database error"))

		err := tc.serviceRepo.CreateSession(ctx, sessionID, userID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_GetSessionByUserID(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()
	ctx := context.Background()
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		tc.mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"session_id", "user_id", "created_at", "expires_at"}).
				AddRow("session-123", userID, now, expiresAt))

		session, err := tc.serviceRepo.GetSessionByUserID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, "session-123", session.SessionID)
		require.Equal(t, userID, session.UserID)
		require.Equal(t, expiresAt, session.ExpiresAt)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		tc.mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		session, err := tc.serviceRepo.GetSessionByUserID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, session)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}

func TestServiceRepo_UpdateSessionExpiry(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest()
	defer tc.teardownTest()
	ctx := context.Background()
	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		tc.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := tc.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.NoError(t, err)
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Session Not Found", func(t *testing.T) {
		tc.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := tc.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Equal(t, "no session found with the provided ID", err.Error())
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("RowsAffected Error", func(t *testing.T) {
		tc.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := tc.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		tc.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnError(errors.New("database error"))

		err := tc.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, tc.mock.ExpectationsWereMet())
	})
}
