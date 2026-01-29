package foodservice

import "net/http"

type Handler interface {
	RegisterUser() http.HandlerFunc
	LoginUser() http.HandlerFunc
	AddNewOrder() http.HandlerFunc
	GetOrdersList() http.HandlerFunc
}
