
# 🚀 Order Service - WB L0

Микросервис для обработки и отображения заказов с использованием Go, PostgreSQL, Redis и Kafka.

## 📚 Оглавление

- [Архитектура](#-архитектура)
- [Быстрый старт](#-быстрый-старт)
- [API Endpoints](#-api-endpoints)
- [Конфигурация](#-конфигурация)
- [Разработка](#-разработка)
- [Генерация тестовых данных](#генерация-тестовых-данных)
- [Тестирование производительности](#-тестирование-производительности)
- [Скрипты для тестирования](#скрипты-для-тестирования)
- [Troubleshooting](#-troubleshooting)
- [Мониторинг состояния сервисов](#мониторинг-состояния-сервисов)
- [Архитектурные особенности](#-архитектурные-особенности)

## 📦 Архитектура

```
order-service/
├── cmd/
│   └── main.go                 # Точка входа
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── order_handler.go # HTTP обработчики
│   │       └── web/
│   │           └── index.html  # Веб-интерфейс
│   ├── repository/
│   │   ├── postgres.go         # PostgreSQL репозиторий
│   │   └── redis.go           # Redis кэш
│   ├── service/
│   │   └── order_service.go    # Бизнес-логика
│   └── kafka/
│       └── kafka_consumer.go   # Kafka consumer
├── domain/
│   └── order.go               # Модели данных
├── scripts/
│   ├── send_orders.go         # Генератор тестовых данных
│   └── performance_benchmark.go # Тест производительности
├── migrations/
│   └── 001_init.up.sql        # Миграции БД
└── docker-compose.yml         # Docker конфигурация
```

## 🚀 Быстрый старт

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd WB_L0_2025
```

### 2. Запуск всех сервисов

```bash
# Запуск всех контейнеров
docker-compose up -d

# Просмотр логов
docker-compose logs -f order-service
```

### 3. Генерация тестовых данных

```bash
# Запуск генератора заказов
docker-compose run --rm order-generator

# Или локально (требуется Go)
go run scripts/send_orders.go
```

### 4. Открытие веб-интерфейса

Откройте в браузере: http://localhost:8081

## 📊 API Endpoints

### Основные методы

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/` | Веб-интерфейс |
| `GET` | `/health` | Health check |
| `GET` | `/order/{order_uid}` | Получить заказ по ID |
| `POST` | `/order` | Создать новый заказ |


### Примеры запросов

**Получить заказ:**
```bash
curl http://localhost:8081/order/test-123456
```

**Создать заказ:**
```bash
curl -X POST http://localhost:8081/order \
  -H "Content-Type: application/json" \
  -d '{
    "order_uid": "test-123456",
    "track_number": "WBILMTESTTRACK",
    "entry": "WBIL",
    "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
    },
    "payment": {
      "transaction": "test-123456",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317
    },
    "items": [
      {
        "chrt_id": 9934930,
        "track_number": "WBILMTESTTRACK",
        "price": 453,
        "rid": "ab4219087a764ae0btest",
        "name": "Mascaras",
        "sale": 30,
        "size": "0",
        "total_price": 317,
        "nm_id": 2389212,
        "brand": "Vivienne Sabo",
        "status": 202
      }
    ]
  }'
```

## ⚙️ Конфигурация

### Переменные окружения

Создайте файл `.env`:

```env
DATABASE_PASSWORD=YOURPASSWORD
DATABASE_NAME=orderservice
USER_NAME=orderuser
APP_PORT=8081
DATABASE_PORT=5432
```

### Порты

| Сервис | Порт | Описание |
|--------|------|----------|
| Order Service | 8081 | Основное API |
| PostgreSQL | 5432 | База данных |
| Kafka | 29092 | Message broker |
| Redis | 6379 | Кэширование |

## 🛠️ Разработка

### Основные команды

```bash
# Запуск всех сервисов
docker-compose up -d

# Остановка всех сервисов
docker-compose down

# Просмотр логов
docker-compose logs -f order-service

# Пересборка и перезапуск
docker-compose down && docker-compose up -d --build
```

## Генерация тестовых данных

```bash
# Запуск генератора заказов через Docker
docker-compose run --rm order-generator

# Запуск локально (требуется Go)
go run scripts/send_orders.go

# Генерация определенного количества заказов
go run scripts/send_orders.go --count 50

# Генерация с указанием Kafka брокеров
go run scripts/send_orders.go --brokers localhost:29092 --count 20
```

## 🚀 Тестирование производительности

### Сравнение скорости Redis vs PostgreSQL

Наш сервис использует кэширование в Redis для значительного ускорения работы:

**Результаты тестирования:**
- **Redis (кэш)**: ~0.01-0.1ms на запрос
- **PostgreSQL (БД)**: ~10-50ms на запрос


### Запуск тестов производительности

```bash
# Тест производительности )
go run scripts/performance_benchmark.go

# Unit-тесты производительности
go test -bench=. -benchmem ./internal/service
```
### Генератор тестовых данных

```bash
# Запуск через Docker
docker-compose run --rm order-generator

# Локальный запуск
go run scripts/send_orders.go

# С параметрами
go run scripts/send_orders.go --brokers localhost:29092 --count 100

# Генерация одного заказа для тестирования
go run scripts/send_orders.go --count 1
```

### Тест производительности

```bash
# Запуск теста производительности
go run scripts/performance_benchmark.go

# Тест показывает реальную разницу между:
# - Кеш (память) = очень быстро
# - БД (вычисления) = медленно
# Результат: кеш в десятки/сотни раз быстрее!
```

## Скрипты для тестирования

```bash
# Бенчмарк производительности сервиса
go test -bench=. ./internal/service

# Тест с измерением памяти
go test -bench=. -benchmem ./internal/service
```

### Мониторинг и дебаг

```bash
# Просмотр логов
docker-compose logs -f order-service
docker-compose logs -f kafka

# Проверка Kafka
docker exec kafka kafka-topics.sh --list --bootstrap-server kafka:29092
docker exec kafka kafka-console-consumer.sh --bootstrap-server kafka:29092 --topic orders --from-beginning

# Проверка Redis
docker exec redis redis-cli keys "*"
docker exec redis redis-cli get "order:test-123456"

# Мониторинг Redis
docker exec redis redis-cli info memory
docker exec redis redis-cli info stats
```

## 🐛 Troubleshooting

### Common Issues

1. **Kafka не доступна**
   ```bash
   docker-compose restart kafka
   docker-compose logs kafka
   ```

2. **Миграции не применились**
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

3. **Redis недоступен**
   ```bash
   docker-compose restart redis
   ```

4. **Порты заняты**
   ```bash
   # Linux/Mac
   lsof -i :8081
   # Windows
   netstat -ano | findstr :8081
   ```

5. **Проблемы с кэшированием**
   ```bash
   # Проверка использования памяти Redis
   docker exec redis redis-cli info memory
   
   # Очистка кэша (если нужно)
   docker exec redis redis-cli flushall
   ```

## Мониторинг состояния сервисов

```bash
# Проверка состояния всех контейнеров
docker-compose ps

# Проверка использования ресурсов
docker stats

# Проверка логов конкретного сервиса
docker-compose logs order-service
```

## 📈 Архитектурные особенности

### Кэширование
- **Redis** используется для кэширования заказов
- Автоматическая очистка при переполнении памяти (`maxmemory-policy allkeys-lru`)
- Лимит памяти: 512MB

### Тестирование производительности:
- **Кеш**: чтение из памяти = ~0.01ms
- **БД**: вычисления и обработка = ~20ms
