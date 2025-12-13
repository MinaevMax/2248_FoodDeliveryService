package http

import (
	"2248_FoodDeliveryService/internal/foodService/mock"
	"2248_FoodDeliveryService/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestHandler_AddNewOrder(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	serviceUC := mock.NewMockUseCase(ctrl)

	handler := NewHandler(serviceUC, logger)
	handlerFunc := handler.AddNewOrder()

	t.Run("AddNewOrder - SUCCESS", func(t *testing.T) {
		userID := uuid.New()
		orderID := uuid.New()

		serviceUC.EXPECT().
			CreateOrder(gomock.Any(), userID.String()).
			Return(orderID, nil)

		order := &models.NewOrderData{
			UserID: userID.String(),
			Amount: 1,
		}

		body, err := json.Marshal(order)
		require.NoError(t, err)
		require.NotNil(t, body)

		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "Application/JSON")
		res := httptest.NewRecorder()

		ctxWithValue := context.WithValue(context.Background(), "userID", userID.String())
		req = req.WithContext(ctxWithValue)

		handlerFunc(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		var response []uuid.UUID
		err = json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response, 1)
		assert.Equal(t, orderID, response[0])
	})

	t.Run("AddNewOrder - UNAUTHORIZED - no userID in context", func(t *testing.T) {
		order := &models.NewOrderData{
			Amount: 1,
		}

		body, _ := json.Marshal(order)

		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "Application/JSON")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), "User not authenticated")
	})

	t.Run("bad request - invalid JSON", func(t *testing.T) {
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader([]byte("{invalid json")))
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Contains(t, res.Body.String(), "Invalid JSON")
	})

	t.Run("bad request - validation error", func(t *testing.T) {
		userID := uuid.New()

		orderData := &models.NewOrderData{}
		body, _ := json.Marshal(orderData)
		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

}

func TestHandler_GetOrdersList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	serviceUC := mock.NewMockUseCase(ctrl)

	handler := NewHandler(serviceUC, logger)
	handlerFunc := handler.GetOrdersList()

	t.Run("success - get all orders (no active param)", func(t *testing.T) {
		userID := uuid.New()
		expectedOrders := []*models.OrderInfo{
			{ID: "order-1", Status: "PACKING", UpdatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: "order-2", Status: "PACKING", UpdatedAt: time.Now().Add(-48 * time.Hour)},
		}

		serviceUC.EXPECT().
			GetOrders(gomock.Any(), userID.String(), false).
			AnyTimes().
			Return(expectedOrders, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/list", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		var response map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)
		require.NotNil(t, response)

		count, ok := response["count"].(float64)
		require.True(t, ok)
		assert.Equal(t, 2, int(count))

		itemsSlice, ok := response["items"].([]interface{})
		require.True(t, ok)
		assert.Len(t, itemsSlice, 2)

		req = httptest.NewRequest(http.MethodGet, "/orders/list?active=false", nil)
		req = req.WithContext(ctx)
		handlerFunc(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("success - get all orders active", func(t *testing.T) {
		userID := uuid.New()
		expectedOrders := []*models.OrderInfo{
			{ID: "order-1", Status: "PACKING", UpdatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: "order-2", Status: "PACKING", UpdatedAt: time.Now().Add(-48 * time.Hour)},
		}

		serviceUC.EXPECT().
			GetOrders(gomock.Any(), userID.String(), true).
			Return(expectedOrders, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/list?active=true", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()
		handlerFunc(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("success - get all orders forbidden", func(t *testing.T) {
		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/orders/list?active=none", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()
		handlerFunc(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("AddNewOrder - UNAUTHORIZED - no userID in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders/list", bytes.NewReader(nil))
		req.Header.Set("Content-Type", "Application/JSON")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Contains(t, res.Body.String(), "User not authenticated")
	})

}

func TestHandler_RegisterUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	serviceUC := mock.NewMockUseCase(ctrl)

	handler := NewHandler(serviceUC, logger)
	handlerFunc := handler.RegisterUser()

	t.Run("error - short login and password", func(t *testing.T) {
		userID := uuid.New()

		body, _ := json.Marshal(map[string]interface{}{
			"login":    "ab",
			"password": "012345",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()
		handlerFunc(res, req)

		req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("{invalid json")))
		handlerFunc(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("error - use case returns nil user with nil error", func(t *testing.T) {
		login := "testuser"
		password := "password123"

		serviceUC.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(nil, errors.New("user already exists"))

		requestBody := `{"login": "testuser", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusConflict, res.Code)
	})

	t.Run("success", func(t *testing.T) {
		login := "testuser"
		password := "password123"
		user := &models.UserData{
			ID: uuid.New().String(),
		}

		serviceUC.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(user, nil)

		requestBody := `{"login": "testuser", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

func TestHandler_LoginUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	serviceUC := mock.NewMockUseCase(ctrl)

	handler := NewHandler(serviceUC, logger)
	handlerFunc := handler.LoginUser()

	t.Run("success - valid credentials with session ID validation", func(t *testing.T) {
		userID := uuid.New()
		login := "user"
		password := "pass"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		email := "a@b.c"

		user := &models.UserData{
			ID:       userID.String(),
			Login:    login,
			Password: string(hashedPassword),
			Email:    email,
		}

		serviceUC.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(user, nil)

		serviceUC.EXPECT().
			CreateSession(gomock.Any(), gomock.AssignableToTypeOf(""), userID.String()).
			Return(nil)

		body := `{"login": "user", "password": "pass"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)

		userData := response["user"].(map[string]interface{})
		assert.Equal(t, userID.String(), userData["id"])
		assert.Equal(t, login, userData["login"])
		assert.Equal(t, email, userData["email"])

		assert.Contains(t, response, "token")
		token, ok := response["token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, token)
	})
}
