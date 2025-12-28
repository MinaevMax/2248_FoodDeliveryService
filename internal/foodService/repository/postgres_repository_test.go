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
	"github.com/stretchr/testify/suite"
)

// TODO поменять на testSuite
type TestSuite struct {
	suite.Suite
	sqlxDB      *sqlx.DB
	mock        sqlmock.Sqlmock
	logger      *slog.Logger
	serviceRepo foodservice.Repository
}

func (suite *TestSuite) SetupTest() {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	suite.sqlxDB = sqlx.NewDb(db, "sqlmock")
	//defer suite.sqlxDB.Close()
	suite.mock = mock
	suite.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	suite.serviceRepo = NewServiceRepo(suite.sqlxDB, nil, suite.logger)
}

func (suite *TestSuite) TeardownTest() {
	suite.sqlxDB.Close()
}

func (suite *TestSuite) TestServiceRepo_CreateOrder() {
	//suite.T().Parallel()

	suite.T().Run("Success", func(t *testing.T) {
		userID := uuid.New()
		orderID := uuid.New()

		suite.mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(orderID))

		createdOrderID, err := suite.serviceRepo.CreateOrder(context.Background(), userID.String())

		require.NoError(t, err)
		require.Equal(t, orderID, createdOrderID)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Error", func(t *testing.T) {
		userID := uuid.New()

		suite.mock.
			ExpectQuery(createOrderQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("abc123"))

		createdOrderID, err := suite.serviceRepo.CreateOrder(context.Background(), userID.String())

		require.Error(t, err)
		require.Equal(t, uuid.UUID{}, createdOrderID)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_ChangeOrderStatus() {
	//suite.T().Parallel()

	orderID := uuid.New()
	status := "ARRIVING"

	order := &models.ChangeOrderStatusData{
		OrderID:   orderID.String(),
		NewStatus: status,
	}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.
			//ExpectPrepare(changeOrderStatusQuery).
			ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
			WithArgs(order.NewStatus, order.OrderID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := suite.serviceRepo.ChangeOrderStatus(context.Background(), order)

		require.NoError(t, err)
	})

	suite.T().Run("Database error", func(t *testing.T) {
		suite.mock.
			ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
			WithArgs(order.NewStatus, orderID).
			WillReturnError(fmt.Errorf("database connection failed"))

		err := suite.serviceRepo.ChangeOrderStatus(context.Background(), order)

		require.Error(t, err)
	})
}

func (suite *TestSuite) TestServiceRepo_GetOrdersForUser() {
	//suite.T().Parallel()

	orderID := uuid.New()
	userID := uuid.New()
	status := "ARRIVING"
	updatedAt := time.Now()

	suite.T().Run("Active Orders", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		suite.mock.
			ExpectQuery(getActiveOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := suite.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), true)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})

	suite.T().Run("All Orders", func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "status", "updated_at"}).
			AddRow(orderID, status, updatedAt)

		var orders []*models.OrderInfo

		orders = append(orders, &models.OrderInfo{orderID.String(), status, updatedAt})

		suite.mock.
			ExpectQuery(getOrdersQuery).
			WithArgs(userID).
			WillReturnRows(rows)

		orders, err := suite.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), false)

		require.NoError(t, err)
		require.NotNil(t, orders)
	})

	suite.T().Run("Database connection timeout", func(t *testing.T) {
		suite.mock.
			ExpectQuery(getActiveOrdersQuery).
			WithArgs(userID.String()).
			WillReturnError(fmt.Errorf("database connection timeout"))

		orders, err := suite.serviceRepo.GetOrdersForUser(context.Background(), userID.String(), true)

		require.Error(t, err)
		require.Nil(t, orders)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_CreateUser() {
	//suite.T().Parallel()

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

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, user.Password, user.Email, user.Phone, user.IsActive, now, now))

		result, err := suite.serviceRepo.CreateUser(ctx, user)

		require.NoError(t, err)
		require.Equal(t, userID.String(), result.ID)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Failed to create user", func(t *testing.T) {
		suite.mock.ExpectQuery(createUserQuery).
			WithArgs(user.Login, user.Password, user.Email, user.Phone, user.IsActive).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, user.Password, user.Email, user.Phone, "not-a-bool", time.Now(), time.Now()))

		result, err := suite.serviceRepo.CreateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_UpdateUser() {
	//suite.T().Parallel()

	ctx := context.Background()
	now := time.Now()
	user := &models.UserData{
		ID:       "user1",
		Login:    "newlogin",
		Email:    "a@b.c",
		Phone:    "+1234567890",
		IsActive: false,
	}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(user.ID, user.Login, "hashedpassword", user.Email, user.Phone, user.IsActive, now.Add(-24*time.Hour), now))

		result, err := suite.serviceRepo.UpdateUser(ctx, user)

		require.NoError(t, err)
		require.Equal(t, user.ID, result.ID)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("User Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(updateUserQuery).
			WithArgs(user.Login, user.Email, user.Phone, user.IsActive, user.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := suite.serviceRepo.UpdateUser(ctx, user)

		require.Error(t, err)
		require.Nil(t, result)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.Contains(t, err.Error(), "no rows in result set")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_GetUserByID() {
	//suite.T().Parallel()

	ctx := context.Background()
	userID := uuid.New()

	suite.T().Run("Success", func(t *testing.T) {
		now := time.Now()
		suite.mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID.String()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "email", "phone", "is_active", "created_at", "updated_at"}).
				AddRow(userID.String(), "user", "pass", "a@b.c", "+1234567890", true, now.Add(-24*time.Hour), now))

		user, err := suite.serviceRepo.GetUserByID(ctx, userID.String())

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, userID.String(), user.ID)
		require.Equal(t, "user", user.Login)
		require.Equal(t, "a@b.c", user.Email)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(getUserByIDQuery).
			WithArgs(userID.String()).
			WillReturnError(sql.ErrNoRows)

		user, err := suite.serviceRepo.GetUserByID(ctx, userID.String())

		require.Error(t, err)
		require.Nil(t, user)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_RegisterUser() {
	//suite.T().Parallel()

	userID := uuid.New()
	login := "user"
	password := "password"

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.
			ExpectQuery(registerUserQuery).
			WithArgs(login, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID.String()))

		user, err := suite.serviceRepo.RegisterUser(context.Background(), login, password)
		require.NoError(t, err)
		require.Equal(t, userID.String(), user.ID)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.mock.
			ExpectQuery(registerUserQuery).
			WithArgs(login, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		user, err := suite.serviceRepo.RegisterUser(context.Background(), login, password)

		require.Error(t, err)
		require.Nil(t, user)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Password too long", func(t *testing.T) {
		password := "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
		_, err := suite.serviceRepo.RegisterUser(context.Background(), login, password)
		require.Error(t, err)
	})
}

func (suite *TestSuite) TestServiceRepo_GetUserByLogin() {
	//suite.T().Parallel()

	ctx := context.Background()
	login := "user"
	userID := uuid.New()
	now := time.Now()

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectQuery(getUserByLoginQuery).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "created_at"}).
				AddRow(userID.String(), login, "hashedpassword", now))

		user, err := suite.serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, userID.String(), user.ID)
		require.Equal(t, login, user.Login)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(sql.ErrNoRows)

		user, err := suite.serviceRepo.GetUserByLogin(ctx, login)

		require.NoError(t, err)
		require.Nil(t, user)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(`SELECT id, login, password, created_at FROM users WHERE login = $1`).
			WithArgs(login).
			WillReturnError(errors.New("database error"))

		user, err := suite.serviceRepo.GetUserByLogin(ctx, login)

		require.Error(t, err)
		require.Nil(t, user)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_DeleteUser() {
	//suite.T().Parallel()

	ctx := context.Background()
	userID := uuid.New()

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := suite.serviceRepo.DeleteUser(ctx, userID.String())

		require.NoError(t, err)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := suite.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Equal(t, "no user found with the provided ID", err.Error())
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnError(errors.New("database error"))

		err := suite.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("RowsAffected Error", func(t *testing.T) {
		suite.mock.ExpectExec(deleteUserQuery).
			WithArgs(userID.String()).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := suite.serviceRepo.DeleteUser(ctx, userID.String())

		require.Error(t, err)
		require.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_CreateSession() {
	//suite.T().Parallel()

	ctx := context.Background()
	sessionID := "session-123"
	userID := "user-123"

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := suite.serviceRepo.CreateSession(ctx, sessionID, userID)

		require.NoError(t, err)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectExec(createSessionQuery).
			WithArgs(sessionID, userID).
			WillReturnError(errors.New("database error"))

		err := suite.serviceRepo.CreateSession(ctx, sessionID, userID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_GetSessionByUserID() {
	//suite.T().Parallel()

	ctx := context.Background()
	userID := "user-123"

	suite.T().Run("Success", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		suite.mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"session_id", "user_id", "created_at", "expires_at"}).
				AddRow("session-123", userID, now, expiresAt))

		session, err := suite.serviceRepo.GetSessionByUserID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, "session-123", session.SessionID)
		require.Equal(t, userID, session.UserID)
		require.Equal(t, expiresAt, session.ExpiresAt)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(getSessionByUserIDQuery).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		session, err := suite.serviceRepo.GetSessionByUserID(ctx, userID)

		require.Error(t, err)
		require.Nil(t, session)
		require.ErrorIs(t, err, sql.ErrNoRows)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *TestSuite) TestServiceRepo_UpdateSessionExpiry() {
	//suite.T().Parallel()

	ctx := context.Background()
	sessionID := "session-123"

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := suite.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.NoError(t, err)
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Session Not Found", func(t *testing.T) {
		suite.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := suite.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Equal(t, "no session found with the provided ID", err.Error())
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("RowsAffected Error", func(t *testing.T) {
		suite.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := suite.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "rows affected error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectExec(updateSessionExpiryQuery).
			WithArgs(sessionID).
			WillReturnError(errors.New("database error"))

		err := suite.serviceRepo.UpdateSessionExpiry(ctx, sessionID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")
		require.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func TestServiceRepoTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
