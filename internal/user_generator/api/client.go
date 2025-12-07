package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"2248_FoodDeliveryService/internal/user_generator/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	log        *slog.Logger
}

func NewClient(baseURL string, log *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// RegisterUser регистрирует пользователя в основном API
func (c *Client) RegisterUser(login, password string) error {
	req := models.RegisterRequest{
		Login:    login,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/auth/register", c.baseURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.log.Info("User registered", slog.String("login", login))
	return nil
}

// LoginUser логинит пользователя и возвращает JWT токен
func (c *Client) LoginUser(login, password string) (string, error) {
	req := models.LoginRequest{
		Login:    login,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/auth/login", c.baseURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to login user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	c.log.Info("User logged in", slog.String("login", login), slog.String("token", loginResp.Token[:20]+"..."))
	return loginResp.Token, nil
}

// CreateOrder создаёт заказ с указанной суммой (требует JWT токен)
func (c *Client) CreateOrder(token string, amount int) (int64, error) {
	req := models.OrderRequest{
		Amount: amount,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/orders/create", c.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create order failed with status %d: %s", resp.StatusCode, string(body))
	}

	var orderID int64
	if err := json.NewDecoder(resp.Body).Decode(&orderID); err != nil {
		return 0, fmt.Errorf("failed to decode order response: %w", err)
	}

	c.log.Info("Order created", slog.Int64("order_id", orderID), slog.Int("amount", amount))
	return orderID, nil
}
