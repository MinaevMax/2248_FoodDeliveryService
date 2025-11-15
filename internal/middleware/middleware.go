package middleware

import (
	"log/slog"
)

type MiddlewareManager struct {
	log          *slog.Logger
}

func NewMiddlewareManager(log *slog.Logger) *MiddlewareManager {
	return &MiddlewareManager{
		log:          log,
	}
}
