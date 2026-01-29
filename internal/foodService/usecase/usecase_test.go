package usecase

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/foodService/mock"
	"2248_FoodDeliveryService/internal/models"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type TestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mock           sqlmock.Sqlmock
	logger         *slog.Logger
	mockerviceRepo *mock.MockRepository
	serviceUC      foodservice.UseCase
}

func (suite *TestSuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())
	suite.ctrl = ctrl
	defer ctrl.Finish()

	logger := slog.Default()
	suite.mockerviceRepo = mock.NewMockRepository(ctrl)
	suite.serviceUC = NewServiceUC(suite.mockerviceRepo, logger)
}

func (suite *TestSuite) TestServiceUC_CreateOrder() {
	suite.T().Parallel()
	userID := uuid.New()
	orderID := uuid.New()

	suite.T().Run("Success", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			CreateOrder(gomock.Any(), userID.String()).
			Return(orderID, nil)

		suite.mockerviceRepo.EXPECT().
			PublishNewOrder(orderID.String()).
			Return(nil).
			Times(1)

		createdOrder, err := suite.serviceUC.CreateOrder(context.Background(), userID.String())
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), createdOrder)
	})

	suite.T().Run("Failed to create", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(uuid.UUID{}, errors.New("database error"))

		suite.mockerviceRepo.EXPECT().
			PublishNewOrder(gomock.Any()).
			Times(0)

		_, err := suite.serviceUC.CreateOrder(context.Background(), "")

		require.Error(suite.T(), err)
	})

	suite.T().Run("Failed to publish", func(t *testing.T) {
		orderID := uuid.New()
		suite.mockerviceRepo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(orderID, nil)
		suite.mockerviceRepo.EXPECT().
			PublishNewOrder(orderID.String()).
			Return(errors.New("rabbitmq connection failed")).
			Times(1)

		_, err := suite.serviceUC.CreateOrder(context.Background(), "")

		require.Error(suite.T(), err)
	})
}

func (suite *TestSuite) TestServiceUC_GetOrders() {
	suite.T().Parallel()

	userID := uuid.New()
	orders := []*models.OrderInfo{}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetOrdersForUser(gomock.Any(), userID.String(), true).
			Return(orders, nil)

		foundOrders, err := suite.serviceUC.GetOrders(context.Background(), userID.String(), true)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), orders, foundOrders)
		require.NotNil(suite.T(), foundOrders)
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetOrdersForUser(gomock.Any(), userID.String(), true).
			Return(orders, errors.New("database error"))

		_, err := suite.serviceUC.GetOrders(context.Background(), userID.String(), true)

		require.Error(suite.T(), err)
	})
}

func (suite *TestSuite) TestServiceUC_GetUserByLogin() {
	suite.T().Parallel()

	userID := uuid.New()
	user := &models.UserData{}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(userID.String())).
			Return(user, nil)

		foundUser, err := suite.serviceUC.GetUserByLogin(context.Background(), userID.String())
		require.NoError(suite.T(), err)
		require.Nil(suite.T(), err)
		require.NotNil(suite.T(), foundUser)
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(userID.String())).
			Return(user, errors.New("database error"))

		_, err := suite.serviceUC.GetUserByLogin(context.Background(), userID.String())

		require.Error(suite.T(), err)
	})
}

func (suite *TestSuite) TestServiceUC_RegisterUser() {
	suite.T().Parallel()

	login := "login"
	password := "pass"

	suite.T().Run("Database error", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(login)).
			Return(nil, errors.New("database error"))

		registeredUser, err := suite.serviceUC.RegisterUser(context.Background(), login, password)
		require.Error(suite.T(), err)
		require.Nil(suite.T(), registeredUser)
	})

	suite.T().Run("User found", func(t *testing.T) {
		existingUser := &models.UserData{}
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(login)).
			Return(existingUser, nil)

		registeredUser, err := suite.serviceUC.RegisterUser(context.Background(), login, password)
		require.Error(suite.T(), err)
		require.Nil(suite.T(), registeredUser)
	})

	suite.T().Run("Success", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(login)).
			Return(nil, nil)

		suite.mockerviceRepo.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(&models.UserData{Login: login, Password: password}, nil).
			Times(1)

		registeredUser, err := suite.serviceUC.RegisterUser(context.Background(), login, password)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), registeredUser)
	})

	suite.T().Run("Registration error", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			GetUserByLogin(gomock.Any(), gomock.Eq(login)).
			Return(nil, nil)

		suite.mockerviceRepo.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(nil, errors.New("database error"))

		registeredUser, err := suite.serviceUC.RegisterUser(context.Background(), login, password)
		require.Error(suite.T(), err)
		require.Nil(suite.T(), registeredUser)
	})
}

func (suite *TestSuite) TestServiceUC_CreateSession() {
	suite.T().Parallel()

	sessionID := uuid.New()
	userID := uuid.New()

	suite.T().Run("Database error", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		err := suite.serviceUC.CreateSession(context.Background(), sessionID.String(), userID.String())
		require.Error(suite.T(), err)
	})

	suite.T().Run("Success", func(t *testing.T) {
		suite.mockerviceRepo.EXPECT().
			CreateSession(gomock.Any(), sessionID.String(), userID.String()).
			Return(nil).
			Times(1)

		err := suite.serviceUC.CreateSession(context.Background(), sessionID.String(), userID.String())
		require.Nil(suite.T(), err)
	})
}

func TestServiceUCTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
