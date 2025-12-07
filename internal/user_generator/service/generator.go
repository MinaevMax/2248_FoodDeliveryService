package service

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	"2248_FoodDeliveryService/internal/user_generator/api"
)

type UserGenerator struct {
	apiClient *api.Client
	log       *slog.Logger
}

func NewUserGenerator(baseURL string, log *slog.Logger) *UserGenerator {
	return &UserGenerator{
		apiClient: api.NewClient(baseURL, log),
		log:       log,
	}
}

// GenerateRandomUser генерирует случайное имя, фамилию и надёжный пароль
func (ug *UserGenerator) GenerateRandomUser() (string, string) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()

	login := fmt.Sprintf("%s_%s_%d", firstName, lastName, rand.Intn(10000))

	password := gofakeit.Password(true, true, true, true, false, 12)

	return login, password
}

// GenerateUser создаёт пользователя, логинится и создаёт 1-3 заказа
func (ug *UserGenerator) GenerateUser() error {
	login, password := ug.GenerateRandomUser()

	// Регистрируем
	err := ug.apiClient.RegisterUser(login, password)
	if err != nil {
		ug.log.Error("failed to register user", slog.String("error", err.Error()))
		return err
	}

	// Логинимся
	token, err := ug.apiClient.LoginUser(login, password)
	if err != nil {
		ug.log.Error("failed to login user", slog.String("error", err.Error()))
		return err
	}

	// Создаем 1-3 случайных заказа
	orderCount := rand.Intn(3) + 1
	for i := 0; i < orderCount; i++ {
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		amount := rand.Intn(500) + 50 // 50-550
		_, err := ug.apiClient.CreateOrder(token, amount)
		if err != nil {
			ug.log.Error("failed to create order", slog.String("error", err.Error()))
		}
	}

	return nil
}

// RunContinuous генерирует пользователей бесконечно с указанным интервалом
func (ug *UserGenerator) RunContinuous(interval time.Duration) {
	ug.log.Info("Starting continuous user generation", slog.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := ug.GenerateUser(); err != nil {
			ug.log.Error("error generating user", slog.String("error", err.Error()))
		}
	}
}

// RunBatch генерирует указанное количество пользователей с интервалом между ними
func (ug *UserGenerator) RunBatch(count int, interval time.Duration) {
	ug.log.Info("Starting batch user generation", slog.Int("count", count), slog.Duration("interval", interval))

	for i := 0; i < count; i++ {
		if err := ug.GenerateUser(); err != nil {
			ug.log.Error("error generating user", slog.String("error", err.Error()))
		}
		if i < count-1 {
			time.Sleep(interval)
		}
	}

	ug.log.Info("Batch generation complete", slog.Int("generated", count))
}
