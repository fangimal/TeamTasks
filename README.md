# TeamTasks

REST API сервис для управления задачами в командах. Текущий этап: управление командами.

## Ожидаемый функционал

REST API для совместной работы над задачами внутри команд.

Планируемые возможности приложения:

- [x] регистрация и вход пользователей;
- [x] JWT-аутентификация и middleware авторизации;
- [x] создание команд;
- [x] ролевая модель `owner`, `admin`, `member`;
- [x] приглашение участников;
- [x] проверка прав доступа к задачам и командам;
- [x] healthcheck через `/health`;
- [ ] управление участниками команды (удаление, смена роли);
- [ ] создание, просмотр, обновление и удаление задач;
- [ ] назначение ответственного за задачу;
- [ ] статусы задач: `todo`, `in_progress`, `done`;
- [ ] фильтрация и пагинация списка задач;
- [ ] история изменений задач;
- [ ] комментарии к задачам;
- [ ] кэширование списков задач через Redis;
- [ ] сложные SQL-запросы для аналитики и выборок по командам;
- [ ] rate limiting и базовая отказоустойчивость;
- [ ] Prometheus-метрики через `/metrics`.

Текущее состояние проекта: реализована инфраструктура приложения, подключение к MySQL, миграции схемы БД, базовый слой Repository, проверка подключения к базе данных в `/health`, регистрация, логин, JWT middleware, создание команд, получение списка команд, приглашение участников с проверкой ролей.

## Требования

- Go 1.26.4
- Docker
- Docker Compose
- Make

## Первый запуск

Создайте или проверьте файл `.env` в корне проекта. Базовые значения уже подготовлены для локальной разработки.

Запуск всех сервисов:

```bash
make up
```

Команда собирает Go-приложение и поднимает три контейнера:

- `app` на порту `8080`
- `mysql` на порту `3306`
- `redis` на порту `6379`

При старте приложение подключается к MySQL, применяет миграции из `migrations/` и завершает запуск с ошибкой, если база данных недоступна.

## Миграции

Применить миграции вручную:

```bash
make migrate-up
```

Откатить миграции:

```bash
make migrate-down
```

Для локального запуска приложения с MySQL и Redis в Docker:

```bash
make local-start
```

Остановить локальную инфраструктуру:

```bash
make local-stop
```

Запуск под Delve для подключения отладчика:

```bash
make local-debug
```

По умолчанию Delve слушает `:40000` в headless-режиме. Если `dlv` не установлен, выполните:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

## Проверка работоспособности

Healthcheck:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-06-21T00:00:00Z"
}
```

Если MySQL недоступен, эндпоинт возвращает `503 Service Unavailable` и `"database": "disconnected"`.

Регистрация пользователя:

```bash
curl -i -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret"}'
```

```pwsh
curl.exe -X POST http://localhost:8080/api/v1/register `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"secret"}'
```

Повторная регистрация с тем же email возвращает `409 Conflict`.

Логин:

```bash
curl -i -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret"}'
```

Ожидаемый ответ содержит JWT:

```json
{
  "token": "...",
  "user": {
    "id": 1,
    "email": "test@test.com"
  }
}
```

Проверка защищенного маршрута:

```bash
TOKEN="<token-from-login>"
curl -i http://localhost:8080/api/v1/ping \
  -H "Authorization: Bearer $TOKEN"
```

Без заголовка `Authorization` или с невалидным токеном маршрут возвращает `401 Unauthorized`.

### Команды

Создание команды (после логина):

```bash
TOKEN="<token-from-login>"
curl -i -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Backend Team"}'
```

Ожидаемый ответ: `201 Created` с данными команды.

Получение списка команд пользователя:

```bash
curl -i http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer $TOKEN"
```

С пагинацией:

```bash
curl -i "http://localhost:8080/api/v1/teams?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

Приглашение пользователя в команду (только `owner` или `admin`):

```bash
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user2@test.com","role":"member"}'
```

Ожидаемый ответ: `200 OK`.

Проверка прав доступа — пользователь с ролью `member` пытается пригласить:

```bash
TOKEN_MEMBER="<token-of-member-user>"
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_MEMBER" \
  -d '{"email":"user3@test.com"}'
```

Ожидаемый ответ: `403 Forbidden` с сообщением "insufficient permissions".

Ошибки валидации:

```bash
# Приглашение несуществующего email
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"nobody@test.com"}'
# 404 Not Found

# Приглашение пользователя, уже состоящего в команде
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user2@test.com"}'
# 409 Conflict
```

Логи приложения:

```bash
make logs
```

Остановка контейнеров:

```bash
make down
```

## Конфигурация

Базовая конфигурация хранится в `configs/config.yaml`. Значения из переменных окружения имеют приоритет над YAML.

Основные переменные окружения:

- `SERVER_HOST`
- `SERVER_PORT`
- `SERVER_READ_TIMEOUT`
- `SERVER_WRITE_TIMEOUT`
- `SERVER_IDLE_TIMEOUT`
- `SERVER_SHUTDOWN_TIMEOUT`
- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_NAME`
- `DATABASE_ROOT_PASSWORD`
- `DATABASE_MAX_OPEN_CONNS`
- `DATABASE_MAX_IDLE_CONNS`
- `DATABASE_CONN_MAX_LIFETIME`
- `REDIS_HOST`
- `REDIS_PORT`
- `LOGGER_LEVEL`
- `LOGGER_FORMAT`
- `JWT_SECRET`
- `JWT_EXPIRATION`

Чтобы изменить порт приложения локально, нужно поменять `SERVER_PORT` в `.env` и перезапустите контейнеры через `make up`.