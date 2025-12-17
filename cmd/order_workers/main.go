package main

import (
	"2248_FoodDeliveryService/internal/config"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"2248_FoodDeliveryService/internal/worker_service/repository"
	"2248_FoodDeliveryService/internal/worker_service/service"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := godotenv.Load(); err != nil {
		log.Error("No .env file found or error loading")
	}

	cfg, err := config.LoadWorkersConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		return
	}

	termCtx, termCancel := context.WithCancel(context.Background())
	go waitSigterm(termCancel, log)

	rabbit, err := rabbitmq.NewRabbitMQ(&cfg.RabbitMQ, termCtx, log)
	if err != nil {
		log.Error("failed to create rabbit connection", slog.Any("error", err))
		return
	}

	orderCh := make(chan string, 100)

	serviceRepo := repository.NewServiceRepo(rabbit, orderCh, log)
	serviceRepo.StartNewOrdersConsumer("new-orders")

	srv := service.NewWorkerPool(cfg.Workers, 100, serviceRepo, log)
	srv.Run(termCtx, orderCh)
	log.Info("service terminated")

}

func waitSigterm(terminate context.CancelFunc, log *slog.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	caughtSignal := <-sigCh

	log.Warn("service starts termination", slog.String("signal", caughtSignal.String()))

	signal.Stop(sigCh)
	close(sigCh)
	terminate()
}
