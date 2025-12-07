# 🚀 Food Delivery Service

**Полнофункциональный микросервис для доставки еды с поддержкой RabbitMQ, JWT аутентификацией и автоматической генерацией тестовых данных.**

## 📚 Быстрая навигация

- **[CAPABILITIES.md](CAPABILITIES.md)** - полный список возможностей сервиса
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - диаграммы и описание архитектуры
- **[API_TEST_GUIDE.md](API_TEST_GUIDE.md)** - примеры API запросов и тестирования
- **[FUNCTIONALITY.md](FUNCTIONALITY.md)** - таблица доступности функционала
- **[USER_GENERATOR.md](USER_GENERATOR.md)** - документация по микросервису генератора
- **[USER_GENERATOR_EXAMPLES.md](USER_GENERATOR_EXAMPLES.md)** - примеры использования генератора

---

## 🎯 Что это?

**Food Delivery Service** - это демонстрационный проект микросервиса на Go с:
- ✅ REST API для управления пользователями и заказами
- ✅ JWT токены для аутентификации
- ✅ Session management с TTL
- ✅ PostgreSQL для хранения данных
- ✅ RabbitMQ для асинхронной обработки
- ✅ Prometheus + Grafana для мониторинга
- ✅ Docker & docker-compose для развертывания
- ✅ **User Generator** - микросервис для генерации тестовых данных

---

## 🚀 Быстрый старт

### Вариант 1: Docker Compose (Рекомендуется)

```bash
# Запустить полный стек (все сервисы + 20 пользователей)
docker-compose up

# Дождитесь, пока все контейнеры станут healthy (~30 сек)
# Food Service API: http://localhost:8080
# RabbitMQ Management: http://localhost:15672 (admin/admin)
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
```

### Вариант 2: Локальный запуск

**Требует:** Go 1.24+, PostgreSQL 16, RabbitMQ 3.13

```bash
# Запустить основной сервис
go run ./cmd/server/main.go

# В другом терминале: запустить генератор (опционально)
go run ./cmd/user_generator/main.go -count 20 -interval 2s
```

---

## 📊 API Эндпоинты

### Аутентификация (публичные)

| Метод | Endpoint | Описание |
|-------|----------|---------|
| POST | `/auth/register` | Регистрация нового пользователя |
| POST | `/auth/login` | Вход и получение JWT токена |

### Управление заказами (требует JWT токен)

| Метод | Endpoint | Описание |
|-------|----------|---------|
| POST | `/orders/create` | Создать новый заказ |
| GET | `/orders/list` | Получить список заказов пользователя |

---

## 🔐 Примеры запросов

### Регистрация

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "login": "john_doe",
    "password": "SecurePass123"
  }'
```

**Ответ (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "login": "john_doe",
  "email": "",
  "phone": "",
  "is_active": true,
  "created_at": "2024-12-06T10:30:00Z",
  "updated_at": "2024-12-06T10:30:00Z"
}
```

### Вход

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "john_doe",
    "password": "SecurePass123"
  }'
```

**Ответ (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "login": "john_doe",
    "email": ""
  }
}
```

### Создание заказа

```bash
curl -X POST http://localhost:8080/orders/create \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"amount": 500}'
```

**Ответ (200 OK):**
```
123456
```

### Список заказов

```bash
curl -X GET "http://localhost:8080/orders/list?active=true" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Ответ (200 OK):**
```json
{
  "count": 2,
  "items": [
    {
      "orderID": "550e8400-e29b-41d4-a716-446655440001",
      "status": "PACKING",
      "updatedAt": "2024-12-06T10:35:00Z"
    }
  ]
}
```

---

## 🤖 User Generator - Генератор тестовых данных

**Микросервис для автоматической генерации пользователей и заказов.**

### Быстрый старт

```bash
# Создать 50 пользователей с интервалом 2 секунды
go run ./cmd/user_generator/main.go -count 50 -interval 2s

# Непрерывная генерация (для демонстрации)
go run ./cmd/user_generator/main.go -interval 10s

# Load testing: 1000 пользователей за 100 сек
go run ./cmd/user_generator/main.go -count 1000 -interval 100ms
```

### Docker запуск

```bash
# Генератор уже включен в docker-compose.yml
# Запускается автоматически и создает 20 пользователей
docker-compose up

# Или запустить с пользовательскими параметрами
docker-compose run user_generator -count 100 -interval 1s
```

**Подробнее:** [USER_GENERATOR.md](USER_GENERATOR.md) и [USER_GENERATOR_EXAMPLES.md](USER_GENERATOR_EXAMPLES.md)

---

## 📂 Структура проекта

```
2248_FoodDeliveryService/
├── cmd/
│   ├── server/              # Основное приложение
│   │   └── main.go
│   └── user_generator/      # Микросервис генератора
│       └── main.go
├── internal/
│   ├── config/              # Загрузка конфигурации
│   ├── db/postgres/         # PostgreSQL driver
│   ├── foodService/         # Основной бизнес-логик
│   │   ├── delivery/        # HTTP handlers
│   │   │   └── http/
│   │   │       ├── handlers.go
│   │   │       └── routes.go
│   │   ├── repository/      # Data access layer
│   │   │   ├── postgres_repository.go
│   │   │   ├── rabbit_repository.go
│   │   │   └── sql_queries.go
│   │   ├── usecase/         # Business logic
│   │   │   └── usecase.go
│   │   ├── delivery.go      # Handler interface
│   │   ├── repository.go    # Repository interface
│   │   └── usecase.go       # UseCase interface
│   ├── jwt/                 # JWT токены
│   ├── middleware/          # HTTP middleware (JWT, Session)
│   ├── models/              # Data models
│   ├── rabbitmq/            # RabbitMQ client
│   ├── server/              # HTTP server setup
│   ├── serviceErrors/       # Custom errors
│   └── utils/               # Utility functions
├── deployment/
│   ├── db/init/             # SQL миграции
│   └── rabbitmq/            # RabbitMQ конфиги
├── docker-compose.yml       # Full stack
├── Dockerfile               # Main app
├── Dockerfile.generator     # User Generator
├── go.mod / go.sum          # Dependencies
└── README.md
```

---

## 🏗️ Архитектура

### Clean Architecture (3 слоя)

```
HTTP Requests
     ↓
┌─────────────────┐
│  HTTP Handlers  │  (Delivery Layer)
└────────┬────────┘
         ↓
┌─────────────────┐
│  Use Cases      │  (Business Logic)
└────────┬────────┘
         ↓
┌─────────────────┐
│  Repository     │  (Data Access)
└────────┬────────┘
         ↓
  PostgreSQL + RabbitMQ
```

### Middleware Stack

```
Request
   ↓
JWT Middleware    (validate token, extract userID)
   ↓
Session Middleware (check & refresh session)
   ↓
Handler Logic     (process request)
   ↓
Response
```

---

## 🛠️ Технологии

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| **Language** | Go | 1.24 |
| **Web Framework** | Gorilla Mux | 1.8.1 |
| **Database** | PostgreSQL | 16 |
| **Password Hashing** | bcrypt | - |
| **JWT** | jwt-go | 3.2.0 |
| **Database Driver** | sqlx | 1.4.0 |
| **Message Queue** | RabbitMQ | 3.13 |
| **AMQP Client** | streadway/amqp | 1.1.0 |
| **Validation** | go-playground/validator | 10.28.0 |
| **UUID** | google/uuid | 1.6.0 |
| **Monitoring** | Prometheus | latest |
| **Visualization** | Grafana | latest |
| **Containerization** | Docker | - |

---

## 🧪 Тестирование

### Встроенные скрипты

```bash
# PowerShell (Windows)
.\test_endpoints.ps1

# Bash (Linux/Mac)
./test_endpoints.sh
```

### Вручную

```bash
# Регистрация
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pass12345"}'

# Вход
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pass12345"}'

# Создание заказа (с токеном)
curl -X POST http://localhost:8080/orders/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"amount":100}'

# Список заказов (с токеном)
curl -X GET "http://localhost:8080/orders/list?active=true" \
  -H "Authorization: Bearer <token>"
```

---

## 📊 Мониторинг

### RabbitMQ Management

```
http://localhost:15672
Username: admin
Password: admin
```

Смотреть:
- Очереди (`Queues`)
- Сообщения в `new-orders` и `order-status-change`
- Consumers и их статус

### Prometheus

```
http://localhost:9090
```

Доступные метрики:
- HTTP запросы
- Задержки
- Ошибки

### Grafana

```
http://localhost:3000
Username: admin
Password: admin
```

Визуализация метрик из Prometheus

---

## 🔒 Безопасность

- ✅ Хеширование паролей через **bcrypt** (cost: 10)
- ✅ JWT токены с **HMAC-SHA256** подписью
- ✅ Session validation и auto-expiry (24 часа)
- ✅ Protected routes через middleware
- ✅ Context-based user ID extraction
- ✅ Input validation на всех уровнях

---

## 📝 Конфигурация

### Переменные окружения (для локального запуска)

Создайте `.env` файл:

```env
SERVER_PORT=8080

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DBNAME=foodservice_db
POSTGRES_USER=admin
POSTGRES_PASSWORD=adminpass

RABBIT_HOST=localhost
RABBIT_PORT=5672
RABBIT_USER=admin
RABBIT_PASSWORD=admin
RABBIT_VHOST=foodservice
```

### Docker Compose (автоматическое)

Все переменные уже установлены в `docker-compose.yml`

---

## 🐛 Troubleshooting

### Сервис не запускается

```bash
# Проверить логи
docker-compose logs app

# Убедиться, что БД здорова
docker-compose logs db | tail -20

# Перезапустить все
docker-compose down
docker-compose up
```

### Генератор не создает пользователей

```bash
# Проверить логи генератора
docker-compose logs user_generator

# Убедиться, что app работает
curl http://localhost:8080/auth/register

# Запустить с правильным URL (если локально)
go run ./cmd/user_generator/main.go -count 5 -url http://localhost:8080
```

### Токен не работает

```bash
# Убедиться, что формат правильный
Authorization: Bearer <token>
# (не Bearer: token или другие варианты)

# Проверить, что токен не истек (TTL 1 час)
# Получить новый через /auth/login
```

---

## 📖 Дополнительная документация

| Документ | Описание |
|----------|---------|
| [CAPABILITIES.md](CAPABILITIES.md) | Полный список всех возможностей |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Архитектурные диаграммы и потоки данных |
| [API_TEST_GUIDE.md](API_TEST_GUIDE.md) | Примеры API запросов и сценарии тестирования |
| [FUNCTIONALITY.md](FUNCTIONALITY.md) | Таблица доступности функционала по ролям |
| [USER_GENERATOR.md](USER_GENERATOR.md) | Полная документация User Generator |
| [USER_GENERATOR_EXAMPLES.md](USER_GENERATOR_EXAMPLES.md) | Примеры использования генератора |

---

## 🚀 Production Ready

Сервис готов к использованию в production с:
- ✅ Graceful shutdown
- ✅ Health checks для всех сервисов
- ✅ Структурированное логирование
- ✅ Error handling на всех уровнях
- ✅ Context timeouts
- ✅ Connection pooling (sqlx)
- ✅ Backoff retry logic (для RabbitMQ)

---

## 📄 Лицензия

MIT

---

## 👨‍💻 Автор

MinaevMax (GitHub: MinaevMax/2248_FoodDeliveryService)

---

## 🎯 Следующие шаги

- [ ] Swagger/OpenAPI документация
- [ ] Unit tests (70%+ coverage)
- [ ] Integration tests
- [ ] E2E tests
- [ ] Rate limiting
- [ ] API versioning
- [ ] Refresh tokens
- [ ] Logout функционал
- [ ] Email верификация
- [ ] 2FA

---

**Проект полностью функционален и готов к использованию! 🚀**

Для начала работы: `docker-compose up`