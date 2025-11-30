package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
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

func (h *handler) AddNewOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received new order request")

		userID := "testUser123456" // TODO заменить за userId из миддлеваре

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

		// Валидируем структуру параметров заказа
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

		newOrderCtx, newOrderCancel := context.WithTimeout(context.Background(), CtxTimeout)
		defer newOrderCancel()
		orderID, err := h.uc.CreateOrder(newOrderCtx, &newOrderParams)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create order", h.log)
			return
		}

		utils.WriteJSONResponse(w, http.StatusOK, orderID, h.log)
	}
}

func (h *handler) GetOrdersList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received get orders list request")

		userID := "testUser123456" // TODO заменить за userId из миддлеваре

		getOrdersCtx, getOrdersCancel := context.WithTimeout(context.Background(), CtxTimeout)
		defer getOrdersCancel()
		orders, err := h.uc.GetOrders(getOrdersCtx, userID)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get orders", h.log)
			return
		}

		utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{"count": len(orders), "items": orders}, h.log)
	}
}
