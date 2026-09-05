# Go Marketplace Microservices

Микросервисный backend маркетплейса на Go. Проект разделен на независимые сервисы с отдельными PostgreSQL базами, синхронным взаимодействием через gRPC и асинхронной обработкой событий через Kafka.

## Стек

- Go 1.25
- PostgreSQL 17
- Apache Kafka
- gRPC
- REST API
- Docker Compose
- golang-migrate
- GitHub Actions CI

## Сервисы

| Сервис | Назначение | Публичный порт | Внутреннее взаимодействие |
| --- | --- | ---: | --- |
| `auth-service` | Регистрация, логин, refresh token, выпуск RSA JWT | `8080` | REST |
| `product-service` | Продавцы, категории, каталог товаров, резервирование остатков | `8081`, `9091` | REST + gRPC |
| `order-service` | Создание заказов, статусы, резервирование товаров, события заказов | `8082` | REST + gRPC-клиент + Kafka |
| `payment-service` | Обработка платежей по событиям заказов | `8083` | Kafka |
| `notification-service` | Асинхронная симуляция уведомлений пользователю | нет | Kafka |

## Поток событий

1. Клиент создает заказ через `order-service`.
2. `order-service` резервирует товар в `product-service` через gRPC.
3. `order-service` сохраняет заказ со статусом `CREATED`.
4. `order-service` публикует `OrderCreatedEvent` в Kafka-топик `orders`.
5. `payment-service` читает `OrderCreatedEvent`, создает платеж, переводит его в `SUCCESS` и публикует `PaymentProcessedEvent` в Kafka-топик `payments`.
6. `order-service` читает `PaymentProcessedEvent` и переводит заказ в статус `PAID`.
7. `notification-service` читает события заказов и платежей, после чего пишет уведомления в stdout.

## gRPC Контракт

Контракт inventory-сервиса лежит здесь:

[proto/product/v1/inventory.proto](proto/product/v1/inventory.proto)

Основные методы:

```proto
service InventoryService {
  rpc ReserveProduct(ReserveProductRequest) returns (ProductResponse);
  rpc ReleaseProduct(ReleaseProductRequest) returns (ProductResponse);
}
```

`order-service` использует этот контракт для резервирования товара и компенсационного возврата остатков. `product-service` дополнительно предоставляет REST-эндпоинты для ручных операций с остатками.

## Топики Kafka

| Топик | Кто публикует | Кто читает | События |
| --- | --- | --- | --- |
| `orders` | `order-service` | `payment-service`, `notification-service` | `OrderCreatedEvent`, `OrderStatusChangedEvent` |
| `payments` | `payment-service` | `order-service`, `notification-service` | `PaymentProcessedEvent` |

События передаются JSON-конвертом:

```json
{
  "type": "OrderCreatedEvent",
  "payload": {},
  "occurred_at": "2026-09-05T12:00:00Z"
}
```

## Локальный запуск

Создайте `.env` в корне репозитория. За основу можно взять [.env.example](.env.example):

```env
AUTH_POSTGRES_USER=user
AUTH_POSTGRES_PASSWORD=pass
AUTH_POSTGRES_DB=authdb
AUTH_DATABASE_DSN=postgres://user:pass@auth-postgres:5432/authdb?sslmode=disable
AUTH_RSA_PRIVATE_KEY_PATH=/app/keys/private.pem

PRODUCT_POSTGRES_USER=user
PRODUCT_POSTGRES_PASSWORD=pass
PRODUCT_POSTGRES_DB=productdb
PRODUCT_DATABASE_DSN=postgres://user:pass@product-postgres:5432/productdb?sslmode=disable

ORDER_POSTGRES_USER=user
ORDER_POSTGRES_PASSWORD=pass
ORDER_POSTGRES_DB=orderdb
ORDER_DATABASE_DSN=postgres://user:pass@order-postgres:5432/orderdb?sslmode=disable

PAYMENT_POSTGRES_USER=user
PAYMENT_POSTGRES_PASSWORD=pass
PAYMENT_POSTGRES_DB=paymentdb
PAYMENT_DATABASE_DSN=postgres://user:pass@payment-postgres:5432/paymentdb?sslmode=disable
```

Для `auth-service` нужен приватный RSA ключ:

```text
auth-service/keys/private.pem
```

Запустить весь стек:

```bash
docker compose up --build
```

Остановить контейнеры:

```bash
docker compose down
```

Остановить контейнеры и удалить локальные volumes баз данных:

```bash
docker compose down -v
```

## REST API

### Auth Service

Базовый URL: `http://localhost:8080`

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/health` | Проверка состояния сервиса |
| `POST` | `/register` | Регистрация пользователя |
| `POST` | `/login` | Авторизация пользователя |
| `POST` | `/refresh-token` | Обновление access token |
| `GET` | `/swagger/` | Swagger UI |

### Product Service

Базовый URL: `http://localhost:8081`

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/health` | Проверка состояния сервиса |
| `POST` | `/api/v1/sellers` | Создать продавца |
| `GET` | `/api/v1/sellers/{id}` | Получить продавца |
| `PUT` | `/api/v1/sellers/{id}` | Обновить продавца |
| `DELETE` | `/api/v1/sellers/{id}` | Удалить продавца |
| `POST` | `/api/v1/categories` | Создать категорию |
| `GET` | `/api/v1/categories` | Получить корневые категории |
| `GET` | `/api/v1/categories/{id}` | Получить категорию |
| `PUT` | `/api/v1/categories/{id}` | Обновить категорию |
| `DELETE` | `/api/v1/categories/{id}` | Удалить категорию |
| `GET` | `/api/v1/categories/{id}/subcategories` | Получить подкатегории |
| `POST` | `/api/v1/products` | Создать товар |
| `GET` | `/api/v1/products/{id}` | Получить товар |
| `PUT` | `/api/v1/products/{id}` | Обновить товар |
| `DELETE` | `/api/v1/products/{id}` | Удалить товар |
| `POST` | `/api/v1/products/{id}/reserve` | Зарезервировать остаток |
| `POST` | `/api/v1/products/{id}/release` | Вернуть остаток |
| `GET` | `/api/v1/sellers/{id}/products` | Получить товары продавца |
| `GET` | `/api/v1/categories/{id}/products` | Получить товары категории |

### Order Service

Базовый URL: `http://localhost:8082`

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/health` | Проверка состояния сервиса |
| `POST` | `/api/v1/orders` | Создать заказ |
| `GET` | `/api/v1/orders/{id}` | Получить заказ |
| `PATCH` | `/api/v1/orders/{id}/status` | Изменить статус заказа |
| `GET` | `/api/v1/users/{id}/orders` | Получить заказы пользователя |

### Payment Service

Базовый URL: `http://localhost:8083`

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/health` | Проверка состояния сервиса |

Обработка платежей работает через Kafka, публичных эндпоинтов для создания платежей нет.

### Notification Service

Публичного HTTP API нет. Сервис читает Kafka-события и логирует симулированные уведомления.

## Примеры запросов

Создать продавца:

```bash
curl -X POST http://localhost:8081/api/v1/sellers \
  -H 'Content-Type: application/json' \
  -d '{"first_name":"Alice","last_name":"Seller","email":"alice@example.com"}'
```

Создать категорию:

```bash
curl -X POST http://localhost:8081/api/v1/categories \
  -H 'Content-Type: application/json' \
  -d '{"name":"Electronics","description":"Devices and accessories"}'
```

Создать товар:

```bash
curl -X POST http://localhost:8081/api/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"seller_id":1,"category_id":1,"name":"Keyboard","description":"Mechanical keyboard","price":19900,"available_quantity":10}'
```

Создать заказ:

```bash
curl -X POST http://localhost:8082/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"items":[{"product_id":1,"quantity":2}]}'
```

Изменить статус заказа вручную:

```bash
curl -X PATCH http://localhost:8082/api/v1/orders/1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"CANCELED"}'
```