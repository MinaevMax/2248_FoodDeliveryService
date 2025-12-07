package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/jwt"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type handler struct {
	log *slog.Logger
	uc  foodservice.UseCase
}

var validate = validator.New()

const CtxTimeout = 5 * time.Second

func NewHandler(uc foodservice.UseCase, log *slog.Logger) foodservice.Handler {
	return &handler{
		log: log,
		uc:  uc,
	}
}

// AddNewOrder обработчик создания нового заказа (защищённый маршрут)
func (h *handler) AddNewOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received new order request")

		userID, ok := r.Context().Value("userID").(string)
		if !ok || userID == "" {
			h.log.Error("User not authenticated, missing userID")
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("failed to read body", slog.Any("error", err))
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to read paged issues body", h.log)
			return
		}
		// Защищаем от утечек
		defer func() {
			if err := r.Body.Close(); err != nil {
				h.log.Error("Failed to close request body", slog.Any("error", err))
			}
		}()
		// Парсим тело в структуру
		newOrderParams := models.NewOrderData{}
		if err := json.Unmarshal(body, &newOrderParams); err != nil {
			utils.WriteJSONError(w, http.StatusBadRequest, "Invalid JSON", h.log)
			return
		}
		newOrderParams.UserID = userID

		var validationError validator.ValidationErrors
		validationErr := validate.Struct(newOrderParams)
		if validationErr != nil {
			errString := ""
			if errors.As(validationErr, &validationError) {
				errString = utils.FormatValidationErrors(validationErr.(validator.ValidationErrors))
			} else {
				errString = validationErr.Error()
			}
			h.log.Warn("got forbidden request params")
			utils.WriteJSONError(w, http.StatusBadRequest, errString, h.log)
			return
		}

		orderIDs := []uuid.UUID{}
		for range newOrderParams.Amount {
			newOrderCtx, newOrderCancel := context.WithTimeout(context.Background(), CtxTimeout)
			defer newOrderCancel()
			orderID, err := h.uc.CreateOrder(newOrderCtx, newOrderParams.UserID)
			if err != nil {
				utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create order", h.log)
				return
			}
			orderIDs = append(orderIDs, orderID)
		}

		utils.WriteJSONResponse(w, http.StatusOK, orderIDs, h.log)
	}
}

// GetOrdersList обработчик получения списка заказов пользователя (защищённый маршрут)
func (h *handler) GetOrdersList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received get orders list request")

		active := r.URL.Query().Get("active")
		isActive := false
		if active != "" {
			switch active {
			case "true":
				isActive = true
			case "false":
				isActive = false
			default:
				utils.WriteJSONError(w, http.StatusBadRequest, "Forbidden order status, should be true/false", h.log)
				return
			}
		}

		userID, ok := r.Context().Value("userID").(string)
		if !ok || userID == "" {
			h.log.Error("User not authenticated, missing userID")
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		getOrdersCtx, getOrdersCancel := context.WithTimeout(context.Background(), CtxTimeout)
		defer getOrdersCancel()
		orders, err := h.uc.GetOrders(getOrdersCtx, userID, isActive)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get orders", h.log)
			return
		}

		utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{"count": len(orders), "items": orders}, h.log)
	}
}

// RegisterUser обработчик регистрации нового пользователя (публичный маршрут)
func (h *handler) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received user registration request")

		ctx, cancel := context.WithTimeout(r.Context(), CtxTimeout)
		defer cancel()

		var req struct {
			Login    string `json:"login" validate:"required,min=3"`
			Password string `json:"password" validate:"required,min=8"`
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("failed to read request body", slog.Any("error", err))
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &req)
		if err != nil {
			h.log.Error("failed to unmarshal request", slog.Any("error", err))
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(req); err != nil {
			h.log.Error("validation error", slog.Any("error", err))
			http.Error(w, "validation failed", http.StatusBadRequest)
			return
		}

		user, err := h.uc.RegisterUser(ctx, req.Login, req.Password)
		if err != nil {
			h.log.Error("failed to register user", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// LoginUser обработчик логина пользователя (публичный маршрут, возвращает JWT токен)
func (h *handler) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received login request")

		ctx, cancel := context.WithTimeout(r.Context(), CtxTimeout)
		defer cancel()

		var req struct {
			Login    string `json:"login" validate:"required"`
			Password string `json:"password" validate:"required"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			h.log.Error("failed to decode request body", slog.Any("error", err))
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(req); err != nil {
			h.log.Error("validation error", slog.Any("error", err))
			http.Error(w, "validation failed", http.StatusBadRequest)
			return
		}

		user, err := h.uc.GetUserByLogin(ctx, req.Login)
		if err != nil || user == nil {
			h.log.Warn("user not found", slog.String("login", req.Login))
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			h.log.Warn("invalid password", slog.String("login", req.Login))
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := jwt.GenerateToken(user.ID, user.Email)
		if err != nil {
			h.log.Error("failed to generate token", slog.Any("error", err))
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		sessionID := uuid.New().String()
		err = h.uc.CreateSession(ctx, sessionID, user.ID)
		if err != nil {
			h.log.Error("failed to create session", slog.Any("error", err))
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":    user.ID,
				"login": user.Login,
				"email": user.Email,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
