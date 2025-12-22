package service

import (
	"2248_FoodDeliveryService/internal/worker_service/mock"
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewWorkerPool(t *testing.T) {
	repo := mock.NewMockRepository(gomock.NewController(t))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	t.Run("should create worker pool with correct configuration", func(t *testing.T) {
		wp := NewWorkerPool(5, 10, repo, logger)

		require.Equal(t, 5, wp.maxWorkers)
		require.Equal(t, 10, wp.maxQueueLen)
		require.Equal(t, 0, int(wp.processed.Load()))
		require.Equal(t, 0, int(wp.enqueued.Load()))
		require.NotNil(t, wp.sem)
		require.NotNil(t, wp.orderQueue)
		require.NotNil(t, wp.rabbitMutex)
	})
}

func TestWorkerPool_ProcessOrder(t *testing.T) {
	repo := mock.NewMockRepository(gomock.NewController(t))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("TestWorkerPool_ProcessOrder - successfully enqueue order", func(t *testing.T) {
		wp := NewWorkerPool(2, 5, repo, logger)
		defer wp.Shutdown()
		repo.EXPECT().
			PublishOrderStatus("order-123", "COMPLETED").
			Return(nil).
			AnyTimes()

		err := wp.ProcessOrder(context.Background(), "order-123")

		require.NoError(t, err)
		require.Equal(t, int64(1), wp.enqueued.Load())
		time.Sleep(2 * time.Second)
	})

	t.Run("TestWorkerPool_ProcessOrder - error when queue is full", func(t *testing.T) {
		t.Skip()
	})

	t.Run("TestWorkerPool_ProcessOrder - handle context cancellation", func(t *testing.T) {
		wp := NewWorkerPool(2, 5, repo, logger)
		defer wp.Shutdown()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		repo.EXPECT().
			PublishOrderStatus(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		err := wp.ProcessOrder(ctx, "order-canceled")

		require.NoError(t, err)
	})
}

func TestWorkerPool_worker(t *testing.T) {
	repo := mock.NewMockRepository(gomock.NewController(t))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	t.Run("TestWorkerPool_ProcessOrder - process orders", func(t *testing.T) {
		repo.EXPECT().
			PublishOrderStatus("test-order-1", "COMPLETED").
			Return(nil).
			Times(1)

		wp := NewWorkerPool(1, 1, repo, logger)

		wp.orderQueue <- "test-order-1"

		time.Sleep(5 * time.Second)

		require.Equal(t, int64(1), wp.processed.Load())
	})

	t.Run("TestWorkerPool_ProcessOrder - order cancelled", func(t *testing.T) {
		t.Skip()
	})

	t.Run("TestWorkerPool_ProcessOrder - handle context timeout", func(t *testing.T) {
		repo.EXPECT().
			PublishOrderStatus("test-order-3", "COMPLETED").
			Return(nil).
			AnyTimes()

		wp := NewWorkerPool(1, 1, repo, logger)
		done := make(chan bool)

		go func() {
			wp.orderQueue <- "test-order-3"
			time.Sleep(6 * time.Second)
			done <- true
		}()

		select {
		case <-done:
			require.True(t, true)
		case <-time.After(7 * time.Second):
			t.Fatal("worker timed out")
		}
	})
}

func TestWorkerPool_Stats(t *testing.T) {
	repo := mock.NewMockRepository(gomock.NewController(t))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	t.Run("should return correct statistics", func(t *testing.T) {
		repo.EXPECT().
			PublishOrderStatus(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		wp := NewWorkerPool(3, 10, repo, logger)
		defer wp.Shutdown()

		for i := 0; i < 3; i++ {
			err := wp.ProcessOrder(context.Background(),
				string(rune('1'+i)))
			require.NoError(t, err)
		}

		time.Sleep(100 * time.Millisecond)

		maxWorkers, currentWorkers, queueLen, processed, enqueued := wp.Stats()

		require.Equal(t, int64(3), maxWorkers)
		require.Equal(t, int64(3), enqueued)
		require.True(t, currentWorkers >= 0 && currentWorkers <= 3)
		require.True(t, queueLen >= 0 && queueLen <= 3)
		require.True(t, processed >= 0 && processed <= 3)
	})
}

func TestWorkerPool_Run(t *testing.T) {
	repo := mock.NewMockRepository(gomock.NewController(t))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("should process orders from external queue", func(t *testing.T) {
		repo.EXPECT().
			PublishOrderStatus("order-1", "COMPLETED").
			Return(nil)
		repo.EXPECT().
			PublishOrderStatus("order-2", "COMPLETED").
			Return(nil)

		wp := NewWorkerPool(2, 10, repo, logger)

		externalQueue := make(chan string, 5)

		externalQueue <- "order-1"
		externalQueue <- "order-2"
		close(externalQueue)

		ctx := context.Background()

		wp.Run(ctx, externalQueue)

		time.Sleep(4 * time.Second)

		require.Equal(t, int64(2), wp.enqueued.Load())
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		wp := NewWorkerPool(2, 10, repo, logger)
		externalQueue := make(chan string, 5)
		ctx, cancel := context.WithCancel(context.Background())

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			wp.Run(ctx, externalQueue)
		}()

		time.Sleep(100 * time.Millisecond)
		cancel()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			_, ok := <-wp.orderQueue
			require.Equal(t, ok, false)
		case <-time.After(1 * time.Second):
			t.Fatal("Run didn't exit on context cancellation")
		}
	})
}

func TestMain(m *testing.M) {
	m.Run()
}
