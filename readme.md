# Marketplace API
## О проекте

Marketplace API — backend-сервис маркетплейса, реализованный на Go с контрактным подходом через OpenAPI.
Проект включает управление товарами, авторизацию пользователей, ролевую модель доступа, работу с заказами и хранение данных в PostgreSQL.

## Основные возможности

- CRUD для товаров `products`
- OpenAPI-спецификация для product API
- кодогенерация server/types из OpenAPI
- PostgreSQL как основная база данных
- миграции через Flyway
- JWT-аутентификация
- refresh token flow
- ролевая модель `USER`, `SELLER`, `ADMIN`
- пагинация и фильтрация списка товаров
- мягкое удаление товаров через статус `ARCHIVED`
- JSON-логирование HTTP-запросов
- `request_id` в middleware и заголовке ответа
- управление заказами и позициями заказа
- транзакционная работа с остатками товаров

## Технологии

- Go
- Chi Router
- PostgreSQL 15
- Flyway
- OpenAPI 3.0.3
- `oapi-codegen`
- JWT
- Docker
- Docker Compose

## Архитектура

- `cmd/main.go` — точка входа приложения
- `openapi/openapi.yaml` — контракт API
- `internal/handlers` — HTTP-слой
- `internal/service` — бизнес-логика
- `internal/repository` — доступ к данным
- `internal/middleware` — auth, request id, logging
- `internal/db` — инициализация подключения к PostgreSQL
- `migrations` — миграции БД

## Модель данных

В проекте используются основные сущности:

- `users`
- `products`
- `orders`
- `order_items`
- `promo_codes`
- `user_operations`
- `refresh_tokens`

Для доменных состояний используются enum-типы:

- статусы товаров
- статусы заказов
- типы скидок промокодов
- типы пользовательских операций

## Аутентификация и роли

Доступные эндпоинты:

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`

После аутентификации пользователь получает:

- `access_token`
- `refresh_token`

JWT содержит:

- `user_id`
- `role`

Поддерживаемые роли:

- `USER`
- `SELLER`
- `ADMIN`

## API товаров

Эндпоинты:

- `POST /products`
- `GET /products`
- `GET /products/{id}`
- `PUT /products/{id}`
- `DELETE /products/{id}`

Поддерживается:

- создание товара
- получение товара по ID
- получение списка товаров
- пагинация через `page` и `size`
- фильтрация по `status`
- фильтрация по `category`
- обновление товара
- мягкое удаление товара

## API заказов

Эндпоинты:

- `POST /orders`
- `GET /orders/{id}`
- `PUT /orders/{id}`
- `POST /orders/{id}/cancel`

Функциональность:

- создание заказа с позициями
- получение заказа по ID
- обновление заказа
- отмена заказа
- резервирование остатков товара
- фиксация snapshot-цены на момент оформления
- расчёт итоговой суммы заказа
- учёт пользовательских операций

## Логирование

Все API-запросы логируются в JSON-формате.
Логи включают:

- `request_id`
- `method`
- `endpoint`
- `status_code`
- `duration_ms`
- `user_id`
- `timestamp`

## Генерация кода из OpenAPI

Установка генератора:

```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

Генерация:

```bash
make generate
```

## Запуск проекта

### Через Docker Compose

```bash
docker compose up --build
```

После запуска сервис доступен по адресу:

```text
http://localhost:8080
```

### Основные компоненты окружения

- `db` — PostgreSQL
- `flyway` — миграции базы данных
- `app` — Go-приложение

## Пример сценария использования

1. Зарегистрировать пользователя через `/auth/register`
2. Выполнить вход через `/auth/login`
3. Создать товар через `/products`
4. Получить список товаров через `GET /products`
5. Создать заказ через `POST /orders`
6. Проверить заказ через `GET /orders/{id}`

## Назначение проекта

Проект демонстрирует реализацию backend-сервиса маркетплейса с OpenAPI-first подходом, CRUD-операциями, авторизацией, ролевым доступом, транзакционной бизнес-логикой и интеграцией с PostgreSQL.
