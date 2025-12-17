package service

import (
    "2248_FoodDeliveryService/internal/worker_service"
    "context"
    "fmt"
    "log/slog"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)

type WorkerPool struct {
    repo        worker_service.Repository
    sem         chan struct{}        // семафор для ограничения параллельности
    orderQueue  chan string          // очередь заказов
    queueWg     sync.WaitGroup       // для graceful shutdown
    maxWorkers  int
    maxQueueLen int
    processed   atomic.Int64
    enqueued    atomic.Int64
	rabbitMutex *sync.Mutex
    log         *slog.Logger
}

func NewWorkerPool(maxWorkers, maxQueueLen int, repository worker_service.Repository, log *slog.Logger) *WorkerPool {
    wp := &WorkerPool{
        repo:        repository,
        sem:         make(chan struct{}, maxWorkers),
        orderQueue:  make(chan string, maxQueueLen),
        maxWorkers:  maxWorkers,
        maxQueueLen: maxQueueLen,
		rabbitMutex: &sync.Mutex{},
        log:         log,
    }
    
    // Запускаем постоянные воркеры
    for i := 0; i < maxWorkers; i++ {
        wp.sem <- struct{}{} // заполняем семафор
        wp.queueWg.Add(1)
        go wp.worker(i)
    }
    
    return wp
}

func (wp *WorkerPool) worker(id int) {
    defer wp.queueWg.Done()
    
    for orderID := range wp.orderQueue {
        sleepTime := time.Duration(rand.Intn(4)+1) * time.Second
        wp.log.Info("Воркер", 
            slog.Int("worker_id", id),
            slog.String("order_id", orderID),
            slog.Duration("sleep_time", sleepTime),
            slog.Int("queue_len", len(wp.orderQueue)))
        
        ctx, cancel := context.WithTimeout(context.Background(), sleepTime)
        orderStatus := "UNDEFINED"
        
        select {
        case <-ctx.Done():
            if ctx.Err() == context.DeadlineExceeded {
                orderStatus = "COMPLETED"
                wp.processed.Add(1)
            } else {
                orderStatus = "CANCELED"
            }
        }
        
        cancel()
        wp.rabbitMutex.Lock()
        if err := wp.repo.PublishOrderStatus(orderID, orderStatus); err != nil {
            wp.log.Error("failed to publish order status",
                slog.String("order_id", orderID),
                slog.String("status", orderStatus),
                slog.Any("error", err))
        }
		wp.rabbitMutex.Unlock()
    }
}

func (wp *WorkerPool) ProcessOrder(ctx context.Context, orderID string) error {
    select {
    case wp.orderQueue <- orderID:
        wp.enqueued.Add(1)
        wp.log.Debug("order enqueued", slog.String("order_id", orderID))
        return nil
    default:
        return fmt.Errorf("queue full, order %s rejected", orderID)
    }
}

func (wp *WorkerPool) Stats() (maxWorkers, currentWorkers, queueLen, processed, enqueued int64) {
    maxWorkers = int64(wp.maxWorkers)
    
    // считаем занятые воркеры
    currentWorkers = 0
    for i := 0; i < wp.maxWorkers; i++ {
        select {
        case <-wp.sem:
        default:
            currentWorkers++
        }
    }
    
    queueLen = int64(len(wp.orderQueue))
    processed = wp.processed.Load()
    enqueued = wp.enqueued.Load()
    
    return
}

func (wp *WorkerPool) Run(ctx context.Context, queue chan string) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case orderID, ok := <-queue:  // ✅ ЧИТАЕМ заказы из внешнего канала!
            if !ok {
                wp.log.Info("order queue closed, shutting down")
                return
            }
            if err := wp.ProcessOrder(ctx, orderID); err != nil {
                wp.log.Error("failed to enqueue order", 
                    slog.String("order_id", orderID), 
                    slog.Any("error", err))
            }
            
        case <-ticker.C:
            max, curr, qLen, proc, enq := wp.Stats()
            wp.log.Info("WorkerPool stats",
                slog.Int64("max_workers", max),
                slog.Int64("current_workers", curr),
                slog.Int64("queue_len", qLen),
                slog.Int64("processed", proc),
                slog.Int64("enqueued", enq))
                
        case <-ctx.Done():
            wp.log.Info("context canceled, shutting down")
            close(wp.orderQueue) // завершаем воркеры
            wp.queueWg.Wait()
            return
        }
    }
}


func (wp *WorkerPool) Shutdown() {
    close(wp.orderQueue)
    wp.queueWg.Wait()
    wp.log.Info("worker pool shutdown complete")
}
