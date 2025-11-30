package server

import (
	"2248_FoodDeliveryService/internal/config"
	serviceHttp "2248_FoodDeliveryService/internal/foodService/delivery/http"
	serviceRepository "2248_FoodDeliveryService/internal/foodService/repository"
	serviceUsecase "2248_FoodDeliveryService/internal/foodService/usecase"
	"2248_FoodDeliveryService/internal/middleware"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	cfg    *config.Config
	db     *sqlx.DB
	rabbit *rabbitmq.RabbitMQ
	log    *slog.Logger
	srv    *http.Server
}

func NewServer(cfg *config.Config, db *sqlx.DB, rabbit *rabbitmq.RabbitMQ, log *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		db:     db,
		rabbit: rabbit,
		log:    log,
	}
}

func (s *Server) Run(errCh chan error) error {
	// Init repositories
	serviceRepo := serviceRepository.NewServiceRepo(s.db, s.rabbit, s.log)

	// Init useCases
	serviceUC := serviceUsecase.NewServiceUC(serviceRepo, s.log)

	// Init handlers
	serviceHandler := serviceHttp.NewHandler(serviceUC, s.log)

	// Init middlewares
	_ = middleware.NewMiddlewareManager(s.log)

	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/orders").Subrouter()

	serviceHttp.MapUserRoutes(apiRouter, serviceHandler)

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Server.Port),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		Handler:      r,
	}

	go func() {
		s.log.Info("Starting server", slog.Int("port", s.cfg.Server.Port))
		err := s.srv.ListenAndServe()
		if err != nil {
			errCh <- fmt.Errorf("listen and server error: %w", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
