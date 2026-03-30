# Gophermart (loyalty system)

Накопительная система лояльности "Гофермарт" — HTTP API на Go + PostgreSQL.

## Эндпоинты

API описано в `SPECIFICATION.md`, а маршруты приложения задаются в `internal/routers/routers.go`.

Без авторизации доступны:
- `POST /api/user/register` — регистрация (автоматическая аутентификация)
- `POST /api/user/login` — логин (выдаётся JWT в Cookie)

Только для авторизованных пользователей:
- `POST /api/user/orders` — загрузка номера заказа (plain text число)
- `GET /api/user/orders` — список загруженных номеров и статусов
- `GET /api/user/balance` — баланс и выведенные за весь период средства
- `POST /api/user/balance/withdraw` — списание баллов за заказ
- `GET /api/user/withdrawals` — история списаний

## Авторизация

Авторизация реализована через Cookie `session_token` (JWT) в `internal/authorization`.

Для подписи JWT обязателен `SECRET_KEY` (переменная окружения).

## Внешний сервис accrual

Начисления по заказам рассчитываются внешним сервисом (black-box).

Приложение обращается к:
- `GET /api/orders/{number}`

Адрес этого сервиса задаётся через `ACCRUAL_SYSTEM_ADDRESS` (или флаг `-r`).

## Конфигурация

Комбинации `ENV`/флагов:
- `RUN_ADDRESS` / `-a` (по умолчанию `:8080`)
- `DATABASE_URI` / `-d` (по умолчанию `host=localhost user=gophermart password=userpassword dbname=gophermart sslmode=disable`)
- `ACCRUAL_SYSTEM_ADDRESS` / `-r` (по умолчанию `http://localhost:8080`)
- `SECRET_KEY` — секрет для JWT (обязательно)

## Запуск

1. Подготовьте PostgreSQL (по умолчанию используется база `gophermart` и пользователь `gophermart`).
2. Соберите бинарник:

```sh
(cd cmd/gophermart && go build -buildvcs=false -o gophermart)
```

3. Запустите:

```sh
export RUN_ADDRESS=":8080"
export DATABASE_URI="host=localhost user=gophermart password=userpassword dbname=gophermart sslmode=disable"
export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8080"
export SECRET_KEY="change-me"

./cmd/gophermart/gophermart
```

Флаги `-a/-d/-r` поддерживаются в `internal/config`.

## Формат запросов

`POST /api/user/register` и `POST /api/user/login`:

```json
{
  "login": "user",
  "password": "pass"
}
```

`POST /api/user/orders`:
- body: `plain text` число (номер заказа)

`POST /api/user/balance/withdraw`:

```json
{
  "order": "2377225624",
  "sum": 751
}
```

