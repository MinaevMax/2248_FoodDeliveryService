package postgres

import (
	"2248_FoodDeliveryService/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
)


func NewPostgresDB(cfg *config.PostgresqlConfig) (*sqlx.DB, error) {
	//Создаем строку для подключения к базе.
	//Значения для нее берутся из переменных, которые мы задали в docker-compose.yml
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Dbname,
	)

	//Подключаемся к базе, и проверяем, что подключение успешно
	db, err := sqlx.Connect(driverName, dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
