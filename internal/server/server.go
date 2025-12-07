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
	serviceRepo.StartStatusChangeConsumer("order-status-change")

	// Init useCases
	serviceUC := serviceUsecase.NewServiceUC(serviceRepo, s.log)

	// Init handlers
	serviceHandler := serviceHttp.NewHandler(serviceUC, s.log)

	// Init middlewares
	middlewareManager := middleware.NewMiddlewareManager(s.log, serviceRepo)

	// Настройка роутеров
	r := mux.NewRouter()
	
	r.Use(middlewareManager.MetricsMiddleware)

	// Применяем миддлвары для защиты маршрутов
	ordersRouter := r.PathPrefix("/orders").Subrouter()
	ordersRouter.Use(middlewareManager.JWTMiddleware)
	ordersRouter.Use(middlewareManager.SessionMiddleware)
	ordersRouter.HandleFunc("/create", serviceHandler.AddNewOrder()).Methods(http.MethodPost)
	ordersRouter.HandleFunc("/list", serviceHandler.GetOrdersList()).Methods(http.MethodGet)

	// Публичные маршруты
	authRouter := r.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", serviceHandler.RegisterUser()).Methods(http.MethodPost)
	authRouter.HandleFunc("/login", serviceHandler.LoginUser()).Methods(http.MethodPost)

	// Создаем сервер
	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Server.Port),
		Handler: r,
	}

	// Запуск сервера в горутине
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
