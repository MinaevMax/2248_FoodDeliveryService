package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	serviceErrors "2248_FoodDeliveryService/internal/serviceErrors"
)

type Config struct {
	Server     ServerConfig
	Postgresql PostgresqlConfig
	RabbitMQ   RabbitMQConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type PostgresqlConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Dbname          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RabbitMQConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Vhost    string	
}

// Загрузка общего конфига
func Load() (Config, error) {
	var cfg Config
	var err error

	cfg.Server, err = LoadServerConfig()
	if err != nil {
		return Config{}, err
	}

	cfg.Postgresql, err = LoadPostgresqlConfig()
	if err != nil {
		return Config{}, err
	}

	cfg.RabbitMQ, err = LoadRabbitConfig()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Загрузка конфига для сервера
func LoadServerConfig() (ServerConfig, error) {
	var cfg ServerConfig
	var err error

	cfg.Port, err = getEnvInt("SERVER_PORT")
	if err != nil {
		return ServerConfig{}, err
	}

	cfg.ReadTimeout, err = getEnvDuration("SERVER_READ_TIMEOUT")
	if err != nil {
		return ServerConfig{}, err
	}

	cfg.WriteTimeout, err = getEnvDuration("SERVER_WRITE_TIMEOUT")
	if err != nil {
		return ServerConfig{}, err
	}

	return cfg, nil
}

// Загрузка конфига для мускуля
func LoadPostgresqlConfig() (PostgresqlConfig, error) {
	cfg := PostgresqlConfig{}
	var err error

	cfg.Host, err = getEnv("POSTGRESQL_HOST")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.Port, err = getEnvInt("POSTGRESQL_PORT")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.User, err = getEnv("POSTGRESQL_USER")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.Password, err = getEnv("POSTGRESQL_PASSWORD")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.Dbname, err = getEnv("POSTGRESQL_DBNAME")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.MaxOpenConns, err = getEnvInt("POSTGRESQL_MAX_OPEN_CONNS")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.MaxIdleConns, err = getEnvInt("POSTGRESQL_MAX_IDLE_CONNS")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.ConnMaxLifetime, err = getEnvDuration("POSTGRESQL_MAX_LIFETIME")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	cfg.ConnMaxIdleTime, err = getEnvDuration("POSTGRESQL_MAX_IDLE_TIME")
	if err != nil {
		return PostgresqlConfig{}, err
	}

	return cfg, nil
}

func LoadRabbitConfig() (RabbitMQConfig, error) {
	cfg := RabbitMQConfig{}
	var err error

	cfg.Host, err = getEnv("RABBIT_HOST")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	cfg.Port, err = getEnvInt("RABBIT_PORT")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	cfg.User, err = getEnv("RABBIT_USER")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	cfg.Password, err = getEnv("RABBIT_PASSWORD")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	cfg.Vhost, err = getEnv("RABBIT_VHOST")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	return cfg, nil
}

// Функция получения из env
func getEnv(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%w: %s", serviceErrors.ErrEnvNotSet, key)
	}
	return v, nil
}

func getEnvInt(key string) (int, error) {
	v, err := getEnv(key)
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("atoi failed: %w", err)
	}
	return intVal, nil
}

func getEnvDuration(key string) (time.Duration, error) {
	v, err := getEnv(key)
	if err != nil {
		return 0, err
	}
	durationVal, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("parse duration failed: %w", err)
	}
	return durationVal, nil
}

func getEnvBool(key string) (bool, error) {
	v, err := getEnv(key)
	if err != nil {
		return false, err
	}
	if v == "true" {
		return true, nil
	}
	if v == "false" {
		return false, nil
	}

	return false, fmt.Errorf("%w: %s", serviceErrors.ErrEnvUnexpectedBool, v)
}

func getEnvSlice(key string) ([]string, error) {
	v, err := getEnv(key)
	if err != nil {
		return nil, err
	}
	var vSlice []string
	err = json.Unmarshal([]byte(v), &vSlice)
	if err != nil {
		return nil, fmt.Errorf("parse slice failed: %w", err)
	}

	return vSlice, nil
}
