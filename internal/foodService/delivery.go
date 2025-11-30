package foodservice

import "net/http"

type Handler interface {
	AddNewOrder() http.HandlerFunc
	GetOrdersList() http.HandlerFunc
}
