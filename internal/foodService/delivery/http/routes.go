package http

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MapUserRoutes регистрирует все маршруты:
// Публичные: POST /auth/register, POST /auth/login
// Защищённые: POST /orders/create, GET /orders/list
func MapUserRoutes(m *mux.Router, h foodservice.Handler) {
	m.HandleFunc("/auth/register", h.RegisterUser()).Methods(http.MethodPost)
	m.HandleFunc("/auth/login", h.LoginUser()).Methods(http.MethodPost)
	m.HandleFunc("/orders/create", h.AddNewOrder()).Methods(http.MethodPost)
	m.HandleFunc("/orders/list", h.GetOrdersList()).Methods(http.MethodGet)
	m.Handle("/metrics", promhttp.Handler())
}
