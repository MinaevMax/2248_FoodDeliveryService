package usecase

import (
	foodservice "2248_FoodDeliveryService/internal/foodService"
	"2248_FoodDeliveryService/internal/models"
	"context"
	"fmt"
	"log/slog"
)

type serviceUC struct {
	serviceRepo foodservice.Repository
	log         *slog.Logger
}

// Новый конструктор UseCase
func NewServiceUC(serviceRepo foodservice.Repository, log *slog.Logger) foodservice.UseCase {
	return &serviceUC{
		serviceRepo: serviceRepo,
		log:         log,
	}
}

// Создание нового заказа
func (uc *serviceUC) CreateOrder(ctx context.Context, orderData *models.NewOrderData) (int64, error) {
	// Создаём заказ в базе данных
	orderId, err := uc.serviceRepo.CreateOrder(ctx, orderData)
	if err != nil {
		// Логируем ошибку создания заказа
		uc.log.Error("Failed to create order", slog.Any("error", err))
		return 0, err
	}

	// Публикуем новый заказ в очередь (если необходимо)
	err = uc.serviceRepo.PublishNewOrder(orderData)
	if err != nil {
		// Логируем ошибку публикации
		uc.log.Error("Failed to publish new order", slog.Any("error", err))
		return 0, err
	}

	// Возвращаем id заказа
	return orderId, nil
}

// Получение списка заказов пользователя
func (uc *serviceUC) GetOrders(ctx context.Context, userID string, isActive bool) ([]*models.OrderInfo, error) {
	// Получаем заказы из репозитория
	orders, err := uc.serviceRepo.GetOrdersForUser(ctx, userID, isActive)
	if err != nil {
		// Логируем ошибку получения заказов
		uc.log.Error("Failed to get orders for user", slog.Any("userID", userID), slog.Any("error", err))
		return nil, err
	}
	return orders, nil
}

// Получение пользователя по логину
func (uc *serviceUC) GetUserByLogin(ctx context.Context, login string) (*models.UserData, error) {
	user, err := uc.serviceRepo.GetUserByLogin(ctx, login)
	if err != nil {
		// Логируем ошибку при получении пользователя
		uc.log.Error("Failed to get user by login", slog.Any("login", login), slog.Any("error", err))
		return nil, err
	}
	return user, nil
}

// Регистрация нового пользователя
func (uc *serviceUC) RegisterUser(ctx context.Context, login, password string) (*models.UserData, error) {
	// Проверка на существование пользователя
	existingUser, err := uc.serviceRepo.GetUserByLogin(ctx, login)
	if err != nil {
		// Если произошла ошибка БД (не "пользователь не найден"), возвращаем ошибку
		uc.log.Error("Failed to check existing user", slog.Any("login", login), slog.Any("error", err))
		return nil, err
	}

	if existingUser != nil {
		// Если пользователь уже существует, возвращаем ошибку
		uc.log.Warn("User already exists", slog.Any("login", login))
		return nil, fmt.Errorf("user already exists")
	}

	// Регистрируем пользователя в репозитории
	user, err := uc.serviceRepo.RegisterUser(ctx, login, password)
	if err != nil {
		// Логируем ошибку регистрации пользователя
		uc.log.Error("Failed to register user", slog.Any("login", login), slog.Any("error", err))
		return nil, err
	}

	// Возвращаем зарегистрированного пользователя
	return user, nil
}

// Создание сессии для пользователя
func (uc *serviceUC) CreateSession(ctx context.Context, sessionID, userID string) error {
	err := uc.serviceRepo.CreateSession(ctx, sessionID, userID)
	if err != nil {
		uc.log.Error("Failed to create session", slog.String("userID", userID), slog.Any("error", err))
		return err
	}
	return nil
}
