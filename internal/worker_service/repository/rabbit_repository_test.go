//go:build integration

package repository

import (
	"2248_FoodDeliveryService/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestServiceRepo_PublishOrderStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	rabbit, logger := utils.SetupRabbit()
	orderCh := make(chan string, 100)

	serviceRepo := NewServiceRepo(rabbit, orderCh, logger)

	t.Run("PublishNewOrder - no error", func(t *testing.T) {
		userID := uuid.New()
		err := serviceRepo.PublishOrderStatus(userID.String(), "PACKING")
		require.NoError(t, err)
	})

	rabbit.Channel.QueuePurge("order-status-change", false)
	time.Sleep(1 * time.Second)
}

func TestServiceRepo_StartNewOrdersConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	rabbit, logger := utils.SetupRabbit()

	t.Run("successfully consumes messages", func(t *testing.T) {
		orderCh := make(chan string, 100)
		serviceRepo := NewServiceRepo(rabbit, orderCh, logger)

		queueName := "new-orders"

		go func() {
			serviceRepo.StartNewOrdersConsumer(queueName)
		}()
		time.Sleep(100 * time.Millisecond)

		orderID := uuid.New()
		body, err := json.Marshal(orderID.String())
		require.NoError(t, err)

		err = rabbit.Channel.Publish(
			"main_exchange", "new-orders", false, false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
			})
		require.NoError(t, err)

		select {
		case consumedOrder := <-orderCh: // Assuming your consumer puts messages here
			require.Equal(t, orderID.String(), consumedOrder)
			t.Logf("Successfully consumed message: %s", consumedOrder)
		case <-time.After(3 * time.Second):
			t.Fatal("Message was not consumed within timeout")
		}
	})
}
