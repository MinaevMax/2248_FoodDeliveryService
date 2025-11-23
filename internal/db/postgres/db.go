package postgres

import (
	"2248_FoodDeliveryService/internal/config"
	"database/sql"
	"fmt"
)

func NewPostgresDB(cfg *config.PostgresqlConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
