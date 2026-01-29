//go:build integration

package repository

import (
	"2248_FoodDeliveryService/internal/models"
	"2248_FoodDeliveryService/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "rabbitmq",
		Tag:          "3-management",
		Env:          []string{},
		ExposedPorts: []string{"5672", "15672"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5672":  {{HostPort: "5672"}},
			"15672": {{HostPort: "15672"}},
		},
	}, func(config *docker.HostConfig) {
		// Auto-remove container on exit
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	pool.MaxWait = 2 * time.Minute
	dsn := fmt.Sprintf("amqp://guest:guest@localhost:%s/", resource.GetPort("5672/tcp"))
	if err := pool.Retry(func() error {
		_, err := amqp.Dial(dsn)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	conn, err := amqp.Dial(dsn)
	defer conn.Close()

	// as of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	m.Run()
}

func TestServiceRepo_PublishNewOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	rabbit, logger := utils.SetupRabbit()

	serviceRepo := NewServiceRepo(nil, rabbit, logger)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		err := serviceRepo.PublishNewOrder(userID.String())
		require.NoError(t, err)
	})

	rabbit.Channel.QueuePurge("new-orders", false)
	time.Sleep(1 * time.Second)
}

func TestServiceRepo_StartStatusChangeConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	rabbit, logger := utils.SetupRabbit()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	serviceRepo := NewServiceRepo(sqlxDB, rabbit, logger)

	orderID := uuid.New()
	testStatus := "processing"
	testUpdatedAt := time.Now().UTC()

	mock.
		ExpectExec(`UPDATE orders SET status = ? WHERE id = ?`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1)).
		WillReturnError(nil)

	t.Run("StartStatusChangeConsumer - no error", func(t *testing.T) {
		queueName := "order-status-change"

		testMessage := models.ChangeOrderStatusData{
			OrderID:   orderID.String(),
			NewStatus: testStatus,
			UpdatedAt: testUpdatedAt,
		}

		messageBody, err := json.Marshal(testMessage)
		require.NoError(t, err)

		done := make(chan bool, 100)
		go func() {
			serviceRepo.StartStatusChangeConsumer(queueName)
			time.Sleep(1 * time.Second)
			done <- true
		}()

		err = rabbit.Channel.Publish(
			"main_exchange",
			queueName,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         messageBody,
				DeliveryMode: amqp.Persistent,
			},
		)
		select {
		case <-done:
			require.NoError(t, mock.ExpectationsWereMet())
			t.Logf("Successfully consumed message: %s", testMessage)
		case <-time.After(3 * time.Second):
			t.Fatal("Message was not consumed within timeout")
		}
		_, err = rabbit.Channel.QueuePurge(queueName, true)
		require.NoError(t, err)
	})
}
