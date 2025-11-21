package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"net/http"

	"github.com/gorilla/mux"
)

func MapUserRoutes(m *mux.Router, h foodservice.Handler) {
	m.HandleFunc("/orders/list", h.GetOrdersList()).Methods(http.MethodGet)
}