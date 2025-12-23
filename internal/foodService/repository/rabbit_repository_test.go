package repository

import (
	"2248_FoodDeliveryService/internal/config"
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/rabbitmq"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func setupRabbit(envPath string) (*config.RabbitMQConfig, context.Context, *slog.Logger) {
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	cfg, err := config.LoadRabbitConfig()
	if err != nil {
		log.Fatalf("failed to load rabbitmq config: %s", err)
	}
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &cfg, ctx, logger
}

//func TestMain(m *testing.M) {
//	pool, err := dockertest.NewPool("")
//	if err != nil {
//		log.Fatalf("Could not construct pool: %s", err)
//	}
//
//	// uses pool to try to connect to Docker
//	err = pool.Client.Ping()
//	if err != nil {
//		log.Fatalf("Could not connect to Docker: %s", err)
//	}
//
//	// pulls an image, creates a container based on it and runs it
//	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
//		Repository:   "rabbitmq",
//		Tag:          "3-management",
//		Mounts:       []string{"/Users/yashar/GolandProjects/2248_FoodDeliveryService/deployment/rabbitmq/definitions.json:/etc/rabbitmq/definitions.json"},
//		Env:          []string{"RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS=-rabbitmq_management load_definitions \"/etc/rabbitmq/definitions.json\""},
//		ExposedPorts: []string{"5672", "15672"},
//		PortBindings: map[docker.Port][]docker.PortBinding{
//			"5672":  {{HostPort: "5672"}},
//			"15672": {{HostPort: "15672"}},
//		},
//	}, func(config *docker.HostConfig) {
//		// Auto-remove container on exit
//		config.AutoRemove = true
//		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
//	})
//
//	if err != nil {
//		log.Fatalf("Could not start resource: %s", err)
//	}
//	pool.MaxWait = 2 * time.Minute
//	dsn := fmt.Sprintf("amqp://admin:admin@localhost:%s/foodservice", resource.GetPort("5672/tcp"))
//	if err := pool.Retry(func() error {
//		_, err := amqp.Dial(dsn)
//		if err != nil {
//			return err
//		}
//		return nil
//	}); err != nil {
//		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
//	}
//
//	conn, err := amqp.Dial(dsn)
//	defer conn.Close()
//
//	// as of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
//	defer func() {
//		if err := pool.Purge(resource); err != nil {
//			log.Fatalf("Could not purge resource: %s", err)
//		}
//	}()
//
//	m.Run() TODO
//}

func TestServiceRepo_PublishNewOrder(t *testing.T) {
	cfg, ctx, logger := setupRabbit("../../../.env")
	rabbit, _ := rabbitmq.NewRabbitMQ(cfg, ctx, logger)

	serviceRepo := NewServiceRepo(nil, rabbit, logger)

	t.Run("PublishNewOrder - no error", func(t *testing.T) {
		userID := uuid.New()
		err := serviceRepo.PublishNewOrder(userID.String())
		require.NoError(t, err)
	})
}

func TestServiceRepo_StartStatusChangeConsumer(t *testing.T) {
	cfg, ctx, logger := setupRabbit("../../../.env")
	rabbit, _ := rabbitmq.NewRabbitMQ(cfg, ctx, logger)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	serviceRepo := NewServiceRepo(sqlxDB, rabbit, logger)

	mock.
		ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1)).
		WillReturnError(nil)

	t.Run("StartStatusChangeConsumer - no error", func(t *testing.T) {
		queueName := "order-status-change"
		orderID := uuid.New()
		testStatus := "processing"
		testUpdatedAt := time.Now().UTC()

		testMessage := models.ChangeOrderStatusData{
			OrderID:   orderID.String(),
			NewStatus: testStatus,
			UpdatedAt: testUpdatedAt,
		}

		messageBody, err := json.Marshal(testMessage)
		require.NoError(t, err)

		done := make(chan bool)
		go func() {
			serviceRepo.StartStatusChangeConsumer(queueName)
			time.Sleep(10 * time.Millisecond)
			done <- true
		}()

		err = rabbit.Channel.Publish(
			"",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         messageBody,
				DeliveryMode: amqp.Persistent,
			},
		)
		require.NoError(t, err)

		select {
		case <-done:
			t.Fatal("Consumer stopped unexpectedly")
		case <-time.After(5 * time.Second):
			break
		}
		_, err = rabbit.Channel.QueuePurge(queueName, true)
		require.NoError(t, err)
	})
}
