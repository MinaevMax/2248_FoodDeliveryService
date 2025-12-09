package usecase

import (
	mocks "2248_FoodDeliveryService/internal/foodService/mock"
	"2248_FoodDeliveryService/internal/models"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//todo покрыть полностью

func TestServiceUC_CreateOrder(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockServiceRepo := mocks.NewMockRepository(ctrl)
	serviceUC := NewServiceUC(mockServiceRepo, logger)

	userID := uuid.New().String()
	orderID := uuid.New()

	mockServiceRepo.EXPECT().
		CreateOrder(gomock.Any(), userID).
		Return(orderID, nil)

	mockServiceRepo.EXPECT().
		PublishNewOrder(userID).
		Return(nil).
		Times(1)

	createdOrder, err := serviceUC.CreateOrder(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)
}

func TestServiceUC_GetOrders(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockServiceRepo := mocks.NewMockRepository(ctrl)
	serviceUC := NewServiceUC(mockServiceRepo, logger)

	userID := uuid.New()

	orders := []*models.OrderInfo{}

	mockServiceRepo.EXPECT().
		GetOrdersForUser(gomock.Any(), userID.String(), true).
		Return(orders, nil)

	foundOrders, err := serviceUC.GetOrders(context.Background(), userID.String(), true)
	require.NoError(t, err)
	require.Equal(t, orders, foundOrders)
	require.NotNil(t, foundOrders)
}

func TestServiceUC_GetUserByLogin(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockServiceRepo := mocks.NewMockRepository(ctrl)
	serviceUC := NewServiceUC(mockServiceRepo, logger)

	userID := uuid.New()
	user := &models.UserData{}

	mockServiceRepo.EXPECT().
		GetUserByLogin(gomock.Any(), gomock.Eq(userID.String())).
		Return(user, nil)

	foundUser, err := serviceUC.GetUserByLogin(context.Background(), userID.String())
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotNil(t, foundUser)
}

func TestServiceUC_RegisterUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockServiceRepo := mocks.NewMockRepository(ctrl)
	serviceUC := NewServiceUC(mockServiceRepo, logger)

	login := "login"
	password := "pass"

	mockServiceRepo.EXPECT().
		GetUserByLogin(gomock.Any(), gomock.Eq(login)).
		Return(nil, sql.ErrNoRows)

	registeredUser, err := serviceUC.RegisterUser(context.Background(), login, password)
	require.Error(t, err)
	require.Nil(t, registeredUser)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestServiceUC_CreateSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockServiceRepo := mocks.NewMockRepository(ctrl)
	serviceUC := NewServiceUC(mockServiceRepo, logger)

	sessionID := uuid.New().String()
	userID := uuid.New().String()

	mockServiceRepo.EXPECT().
		CreateSession(gomock.Any(), sessionID, userID).
		Return(nil).
		Times(1)

	err := serviceUC.CreateSession(context.Background(), sessionID, userID)
	require.Nil(t, err)
}
