**Food Delivery Service** - это демонстрационный проект микросервиса на Go с:
- REST API для управления пользователями и заказами
- JWT токены для аутентификации
- Session management с TTL
- PostgreSQL для хранения данных
- RabbitMQ для асинхронной обработки
- Prometheus + Grafana для мониторинга
- Docker & docker-compose для развертывания
- **User Generator** - микросервис для генерации тестовых данных

---

## Быстрый старт

### Вариант 1: Docker Compose (Рекомендуется)

```bash
# Запустить полный стек (все сервисы + 20 пользователей)
docker compose up

# Дождитесь, пока все контейнеры станут healthy (~30 сек)
# Food Service API: http://localhost:8080
# RabbitMQ Management: http://localhost:15672 (admin/admin)
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
```


##  API Эндпоинты

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
