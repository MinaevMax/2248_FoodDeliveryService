package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/utils"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type handler struct {
	log *slog.Logger
	uc  foodservice.UseCase
}

var validate = validator.New()

func NewHandler(uc foodservice.UseCase, log *slog.Logger) foodservice.Handler {
	return &handler{
		log: log,
		uc:  uc,
	}
}

func (h *handler) GetOrdersList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received orders list request")

		// TODO тут должна быть обработка доступов/прав пользоваетеля
	}
}

func (h *handler) AddNewOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Received new order request")

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

		// TODO тут должна быть работа по созданию заказа
	}
}
