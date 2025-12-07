package main

import (
	"flag"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"2248_FoodDeliveryService/internal/user_generator/service"
)

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the food service API")
	count := flag.Int("count", 0, "Number of users to generate (0 = continuous)")
	interval := flag.Duration("interval", 5*time.Second, "Interval between user generations")

	flag.Parse()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create generator
	generator := service.NewUserGenerator(*baseURL, logger)

	if *count > 0 {
		// Batch mode
		generator.RunBatch(*count, *interval)
	} else {
		// Continuous mode
		generator.RunContinuous(*interval)
	}
}
