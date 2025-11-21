package foodservice

import "net/http"

type Handler interface {
	GetOrdersList() http.HandlerFunc
}
