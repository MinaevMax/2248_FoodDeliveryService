package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// todo поменять на testsuite
type testContext struct {
	ctrl      *gomock.Controller
	serviceUC *mock.MockUseCase
	handler   foodservice.Handler
}

func (c *testContext) setupTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	c.ctrl = ctrl
	defer ctrl.Finish()

	logger := slog.Default()
	c.serviceUC = mock.NewMockUseCase(ctrl)
	c.handler = NewHandler(c.serviceUC, logger)
}

func (c *testContext) teardownTest() {
	c.ctrl.Finish()
}

func TestHandler_AddNewOrder(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest(t)
	defer tc.teardownTest()

	handlerFunc := tc.handler.AddNewOrder()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		orderID := uuid.New()

		tc.serviceUC.EXPECT().
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

		require.Equal(t, http.StatusOK, res.Code)

		var response []uuid.UUID
		err = json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response, 1)
		require.Equal(t, orderID, response[0])
	})

	t.Run("Unauthorized", func(t *testing.T) {
		order := &models.NewOrderData{
			Amount: 1,
		}

		body, _ := json.Marshal(order)

		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "Application/JSON")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Contains(t, res.Body.String(), "User not authenticated")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader([]byte("{invalid json")))
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Contains(t, res.Body.String(), "Invalid JSON")
	})

	t.Run("Validation error", func(t *testing.T) {
		orderData := &models.NewOrderData{}
		body, _ := json.Marshal(orderData)
		req := httptest.NewRequest(http.MethodPost, "/orders/create", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusBadRequest, res.Code)
	})
}

func TestHandler_GetOrdersList(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest(t)
	defer tc.teardownTest()

	handlerFunc := tc.handler.GetOrdersList()

	expectedOrders := []*models.OrderInfo{
		{ID: "order-1", Status: "PACKING", UpdatedAt: time.Now().Add(-24 * time.Hour)},
		{ID: "order-2", Status: "PACKING", UpdatedAt: time.Now().Add(-48 * time.Hour)},
	}

	userID := uuid.New()

	t.Run("Get all orders", func(t *testing.T) {
		tc.serviceUC.EXPECT().
			GetOrders(gomock.Any(), userID.String(), false).
			AnyTimes().
			Return(expectedOrders, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/list", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusOK, res.Code)

		var response map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)
		require.NotNil(t, response)

		count, ok := response["count"].(float64)
		require.True(t, ok)
		require.Equal(t, 2, int(count))

		itemsSlice, ok := response["items"].([]interface{})
		require.True(t, ok)
		require.Len(t, itemsSlice, 2)

		req = httptest.NewRequest(http.MethodGet, "/orders/list?active=false", nil)
		req = req.WithContext(ctx)
		handlerFunc(res, req)
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("Get all active orders", func(t *testing.T) {
		tc.serviceUC.EXPECT().
			GetOrders(gomock.Any(), userID.String(), true).
			Return(expectedOrders, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/list?active=true", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()
		handlerFunc(res, req)
		require.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("Get all orders with invalid param", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/orders/list?active=none", nil)
		ctx := context.WithValue(req.Context(), "userID", userID.String())
		req = req.WithContext(ctx)
		res := httptest.NewRecorder()
		handlerFunc(res, req)
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders/list", bytes.NewReader(nil))
		req.Header.Set("Content-Type", "Application/JSON")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Contains(t, res.Body.String(), "User not authenticated")
	})

}

func TestHandler_RegisterUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest(t)
	defer tc.teardownTest()

	handlerFunc := tc.handler.RegisterUser()
	userID := uuid.New()

	t.Run("Short login and password", func(t *testing.T) {

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
		require.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("User already exists", func(t *testing.T) {
		login := "testuser"
		password := "password123"

		tc.serviceUC.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(nil, errors.New("user already exists"))

		requestBody := `{"login": "testuser", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusConflict, res.Code)
	})

	t.Run("Success", func(t *testing.T) {
		login := "testuser"
		password := "password123"
		user := &models.UserData{
			ID: uuid.New().String(),
		}

		tc.serviceUC.EXPECT().
			RegisterUser(gomock.Any(), login, password).
			Return(user, nil)

		requestBody := `{"login": "testuser", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusCreated, res.Code)
	})
}

func TestHandler_LoginUser(t *testing.T) {
	t.Parallel()
	tc := &testContext{}
	tc.setupTest(t)
	defer tc.teardownTest()

	handlerFunc := tc.handler.LoginUser()

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

	t.Run("Success", func(t *testing.T) {
		tc.serviceUC.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(user, nil)

		tc.serviceUC.EXPECT().
			CreateSession(gomock.Any(), gomock.AssignableToTypeOf(""), userID.String()).
			Return(nil)

		body := `{"login": "user", "password": "pass"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		handlerFunc(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "application/json", res.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &response)
		require.NoError(t, err)

		userData := response["user"].(map[string]interface{})
		require.Equal(t, userID.String(), userData["id"])
		require.Equal(t, login, userData["login"])
		require.Equal(t, email, userData["email"])

		require.Contains(t, response, "token")
		token, ok := response["token"].(string)
		require.True(t, ok)
		require.NotEmpty(t, token)
	})

	t.Run("Invalid json body", func(t *testing.T) {
		t.Parallel()

		reqBody := strings.NewReader(`{"login": "user", "password": "pass"`) // missing closing brace
		req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
		w := httptest.NewRecorder()

		handlerFunc(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "invalid request")
	})

	t.Run("Validation failed", func(t *testing.T) {
		t.Parallel()

		reqBody := strings.NewReader(`{"login": "", "password": "password123"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
		w := httptest.NewRecorder()

		handlerFunc(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "validation failed")
	})

	t.Run("User not found", func(t *testing.T) {
		t.Parallel()

		reqBody := strings.NewReader(`{"login": "nonexistent", "password": "password123"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
		w := httptest.NewRecorder()

		tc.serviceUC.EXPECT().GetUserByLogin(gomock.Any(), "nonexistent").
			Return(nil, errors.New("user not found"))

		handlerFunc(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "invalid credentials")
	})

	t.Run("Invalid password", func(t *testing.T) {
		t.Parallel()

		reqBody := strings.NewReader(`{"login": "user", "password": "wrongpassword"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
		w := httptest.NewRecorder()

		tc.serviceUC.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(user, nil)

		handlerFunc(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Body.String(), "invalid credentials")
	})

	t.Run("Session creation failed", func(t *testing.T) {
		t.Parallel()

		reqBody := strings.NewReader(`{"login": "testuser", "password": "pass"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
		w := httptest.NewRecorder()

		tc.serviceUC.EXPECT().GetUserByLogin(gomock.Any(), "testuser").
			Return(user, nil)

		tc.serviceUC.EXPECT().CreateSession(gomock.Any(), gomock.Any(), userID.String()).
			Return(errors.New("database error"))

		handlerFunc(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "failed to create session")
	})
}
