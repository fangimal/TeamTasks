# TeamTasks

REST API сервис для управления задачами в командах с ролевой моделью, аудитом изменений, кэшированием и мониторингом.

**Стек:** Go 1.26.4, MySQL 9.7, Redis 8.8, Prometheus 3.3, Docker Compose.

---

## Архитектура

Проект следует **Clean Architecture** с разделением на слои:

- `domain` — сущности, интерфейсы репозиториев, ошибки
- `usecase` — бизнес-логика (зависит только от интерфейсов domain)
- `repository/mysql` — реализация репозиториев для MySQL
- `repository/redis` — реализация кэша в Redis
- `delivery/http` — HTTP-обработчики
- `delivery/middleware` — Auth, Rate Limit, Logging
- `delivery/router` — настройка маршрутов (`net/http`)
- `config` — загрузка конфигурации (YAML + ENV)
- `pkg/jwt`, `pkg/password`, `pkg/response` — утилиты

```
cmd/
  app/main.go          # Точка входа
  migrate/main.go      # CLI для миграций
internal/
  config/              # Конфигурация
  domain/              # Модели, интерфейсы, ошибки
  usecase/             # Бизнес-логика
  delivery/
    http/              # Хендлеры
    middleware/        # Auth, Rate Limit, Logging
    router/            # Маршруты
  repository/
    mysql/             # Репозитории MySQL
    redis/             # Репозиторий Redis (кэш)
  external/email/      # Circuit Breaker для email
  monitoring/          # Prometheus метрики
  logger/              # structured slog
pkg/                   # JWT, password, response
migrations/            # SQL-миграции
configs/config.yaml    # Базовая конфигурация
prometheus/            # Конфиг Prometheus
```

---

## Быстрый старт

```bash
# 1. Клонировать репозиторий
git clone https://github.com/fangimal/TeamTasks.git
cd TeamTasks

# 2. Скопировать .env.example в .env
cp .env.example .env

# 3. Запустить проект
make up
```

Все сервисы поднимаются одной командой:

| Сервис | Порт | Описание |
|--------|------|----------|
| `app` | 8080 | HTTP API |
| `mysql` | 3306 | База данных |
| `redis` | 6379 | Кэш + rate limiter |
| `prometheus` | 9090 | Сбор метрик |

После запуска проверить работоспособность:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ: `{"status":"ok","database":"connected","cache":"connected","timestamp":"..."}`

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `SERVER_HOST` | `0.0.0.0` | Хост HTTP-сервера |
| `SERVER_PORT` | `8080` | Порт HTTP-сервера |
| `SERVER_READ_TIMEOUT` | `10s` | Таймаут чтения запроса |
| `SERVER_WRITE_TIMEOUT` | `10s` | Таймаут записи ответа |
| `SERVER_IDLE_TIMEOUT` | `60s` | Таймаут бездействия соединения |
| `SERVER_SHUTDOWN_TIMEOUT` | `5s` | Таймаут graceful shutdown |
| `DATABASE_HOST` | `mysql` | Хост MySQL |
| `DATABASE_PORT` | `3306` | Порт MySQL |
| `DATABASE_USER` | `teamtasks` | Пользователь MySQL |
| `DATABASE_PASSWORD` | `teamtasks` | Пароль MySQL |
| `DATABASE_NAME` | `teamtasks` | Название БД |
| `DATABASE_ROOT_PASSWORD` | `root` | Root-пароль MySQL |
| `DATABASE_MAX_OPEN_CONNS` | `25` | Макс. открытых соединений |
| `DATABASE_MAX_IDLE_CONNS` | `5` | Макс. бездействующих соединений |
| `DATABASE_CONN_MAX_LIFETIME` | `5m` | Время жизни соединения |
| `REDIS_HOST` | `redis` | Хост Redis |
| `REDIS_PORT` | `6379` | Порт Redis |
| `REDIS_PASSWORD` | `` | Пароль Redis |
| `REDIS_DB` | `0` | Номер БД Redis |
| `LOGGER_LEVEL` | `info` | Уровень лога (debug/info/warn/error) |
| `LOGGER_FORMAT` | `json` | Формат лога (json/text) |
| `JWT_SECRET` | `change-me-in-production` | Секрет для подписи JWT |
| `JWT_EXPIRATION` | `24h` | Время жизни JWT-токена |
| `RATE_LIMIT_ENABLED` | `true` | Включить rate limiter |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | `100` | Лимит запросов в минуту |
| `EMAIL_SERVICE_URL` | `https://httpbin.org/post` | URL email-сервиса |
| `EMAIL_SERVICE_TIMEOUT` | `5s` | Таймаут запроса к email |

---

## API Endpoints

### Публичные (без JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/health` | Проверка работоспособности |
| `GET` | `/metrics` | Prometheus метрики |
| `POST` | `/api/v1/register` | Регистрация пользователя |
| `POST` | `/api/v1/login` | Вход в систему |

### Защищённые (требуют JWT + rate limit)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/v1/ping` | Проверка авторизации |
| `POST` | `/api/v1/teams` | Создание команды |
| `GET` | `/api/v1/teams` | Список команд пользователя |
| `POST` | `/api/v1/teams/{id}/invite` | Приглашение в команду |
| `POST` | `/api/v1/tasks` | Создание задачи |
| `GET` | `/api/v1/tasks` | Список задач (фильтрация + пагинация) |
| `GET` | `/api/v1/tasks/{id}` | Получение задачи |
| `PUT` | `/api/v1/tasks/{id}` | Обновление задачи |
| `GET` | `/api/v1/tasks/{id}/history` | История изменений задачи |
| `POST` | `/api/v1/tasks/{id}/comments` | Добавить комментарий |
| `GET` | `/api/v1/tasks/{id}/comments` | Список комментариев |
| `GET` | `/api/v1/analytics/team-stats` | Статистика команд |
| `GET` | `/api/v1/analytics/top-users` | Топ пользователей |
| `GET` | `/api/v1/analytics/integrity-check` | Проверка целостности |

### Параметры запроса списка задач

`GET /api/v1/tasks` поддерживает:

- `team_id` (обязательный) — ID команды
- `status` — фильтр по статусу (`todo`, `in_progress`, `done`)
- `assignee_id` — фильтр по исполнителю
- `limit` — количество записей (по умолчанию 10)
- `offset` — смещение (по умолчанию 0)

---

## Разработка и тестирование

```bash
# Unit-тесты (без Docker)
make test-unit

# Интеграционные тесты (требуют Docker)
make test-integration

# Все тесты
make test

# Покрытие кода
make test-coverage

# Линтинг (требует golangci-lint)
make lint

# Форматирование кода
make fmt
```

### Локальный запуск без Docker

```bash
# Поднять только MySQL и Redis
make infra-up

# Запустить приложение локально
make run-local   # или make local-start
```

---

## Мониторинг (Prometheus)

Эндпоинт `/metrics` доступен на порту приложения.

### Доступные метрики

- `teamtasks_http_requests_total` — количество запросов (labels: `method`, `path`, `status`)
- `teamtasks_http_request_duration_seconds` — гистограмма длительности
- `teamtasks_db_max_open_connections`
- `teamtasks_db_open_connections`
- `teamtasks_db_in_use_connections`
- `teamtasks_db_idle_connections`
- `teamtasks_db_wait_count_total`
- `teamtasks_db_wait_duration_seconds_total`

### Prometheus UI

После `make up` откройте http://localhost:9090.

Пример PromQL запроса:

```promql
rate(teamtasks_http_requests_total[1m])
```

---

## Команды Makefile

| Цель | Описание |
|------|----------|
| `build` | Сборка Docker-образов |
| `up` | Запуск всех сервисов |
| `down` | Остановка всех сервисов |
| `logs` | Логи приложения |
| `infra-up` | Запуск MySQL + Redis |
| `local-start` | Инфраструктура + локальный запуск |
| `local-stop` | Остановка инфраструктуры |
| `run-local` | Локальный запуск приложения |
| `local-debug` | Запуск под Delve |
| `migrate-up` | Применить миграции |
| `migrate-down` | Откатить миграции |
| `test` | Все тесты |
| `test-unit` | Unit-тесты |
| `test-integration` | Интеграционные тесты |
| `test-coverage` | Покрытие кода |
| `lint` | Линтинг |
| `fmt` | Форматирование |
| `clean` | Остановка + удаление томов |

---

---

## Проверка соответствия ТЗ (end-to-end)

### 1. Регистрация и аутентификация

```bash
# Регистрация
curl -i -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret"}'

# Повторная регистрация — 409 Conflict
curl -i -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret"}'

# Логин (получение JWT)
curl -i -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret"}'
```

Ожидаемый ответ логина: `200 OK` с `{"token":"...","user":{"id":1,"email":"test@test.com"}}`.

Фиксация токена в переменную (PowerShell):

```bash
$TOKEN = (curl -s -X POST http://localhost:8080/api/v1/login `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"secret"}' | `
  ConvertFrom-Json | Select -ExpandProperty token)
```

Проверка валидности JWT:

```bash
curl -i http://localhost:8080/api/v1/ping -H "Authorization: Bearer $TOKEN"
# 200 OK

curl -i http://localhost:8080/api/v1/ping
# 401 Unauthorized — без токена
```

### 2. Управление командами

```bash
# Создание команды (создатель становится owner)
curl -i -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Backend Team"}'

# Список команд пользователя (с пагинацией)
curl -i "http://localhost:8080/api/v1/teams?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Приглашение пользователя в команду (только owner/admin)
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user2@test.com","role":"member"}'
# 200 OK

# Проверка RBAC — member не может приглашать
# (залогиниться как user2, попробовать пригласить)
# 403 Forbidden
```

### 3. Управление задачами

```bash
# Создание задачи (только член команды)
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Implement auth","description":"Add JWT","assignee_id":1,"team_id":1}'
# 201 Created

# Попытка создать задачу в чужой команде — 403
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Hack","assignee_id":1,"team_id":999}'

# Назначение исполнителя не из команды — 400
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Bad","assignee_id":999,"team_id":1}'

# Получение списка задач с фильтрацией и пагинацией
curl -i "http://localhost:8080/api/v1/tasks?team_id=1&status=todo&limit=5&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Получение задачи по ID
curl -i http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer $TOKEN"

# Обновление задачи (с записью в историю)
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Updated title","status":"in_progress"}'
# 200 OK + автоматическая запись в task_history

# Проверка RBAC — member не может обновить чужую задачу
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_MEMBER" \
  -d '{"title":"Hacked"}'
# 403 Forbidden

# История изменений задачи
curl -i "http://localhost:8080/api/v1/tasks/1/history?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"
# Ожидаемый ответ: массив записей с old_value/new_value (JSON)

# Комментарии к задаче
curl -i -X POST http://localhost:8080/api/v1/tasks/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"text":"Great job!"}'
# 201 Created

# Получение комментариев с пагинацией
curl -i "http://localhost:8080/api/v1/tasks/1/comments?limit=5&offset=0" \
  -H "Authorization: Bearer $TOKEN"
# Ожидаемый ответ: {"data":[...],"total":1,"limit":5,"offset":0}
```

### 4. Сложные SQL-запросы

#### а) JOIN 3+ таблиц + агрегация — статистика команд

Запрос объединяет `teams`, `team_members`, `tasks`; вычисляет количество участников и выполненных задач за 7 дней.

```bash
curl -i http://localhost:8080/api/v1/analytics/team-stats \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK`:
```json
[
  {
    "team_id": 1,
    "team_name": "Backend Team",
    "member_count": 3,
    "done_tasks_last_7_days": 5
  }
]
```

#### б) Оконная функция DENSE_RANK — топ-3 пользователей

Использует CTE + `DENSE_RANK() OVER (PARTITION BY team_id ORDER BY task_count DESC)`.

```bash
curl -i http://localhost:8080/api/v1/analytics/top-users \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK`:
```json
[
  {
    "team_id": 1,
    "user_id": 1,
    "user_email": "test@test.com",
    "task_count": 10,
    "rank": 1
  }
]
```

#### в) Проверка целостности данных

Находит задачи, где `assignee_id` не состоит в `team_members` команды задачи.

```bash
curl -i http://localhost:8080/api/v1/analytics/integrity-check \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK` (пустой массив = нарушений нет):
```json
[]
```

### 5. Кэширование Redis

Список задач команды кэшируется в Redis с TTL 5 минут. Ключ строится из параметров фильтра (`team_id`, `status`, `assignee_id`) и пагинации.

```bash
# Первый запрос — Cache Miss (данные из MySQL)
time curl -s "http://localhost:8080/api/v1/tasks?team_id=1" \
  -H "Authorization: Bearer $TOKEN"

# Второй запрос — Cache Hit (данные из Redis, быстрее)
time curl -s "http://localhost:8080/api/v1/tasks?team_id=1" \
  -H "Authorization: Bearer $TOKEN"

# Инвалидация кэша при обновлении задачи
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"After cache"}'
# Следующий GET запрос снова Cache Miss

# Проверка ключей в Redis
docker-compose exec redis redis-cli keys 'tasks:team:*'
```

### 6. Rate Limiting

На защищённые эндпоинты наложен лимит 100 запросов/мин на пользователя (Redis INCR/EXPIRE).

```bash
# Серия быстрых запросов
for ($i=0; $i -le 100; $i++) {
  curl -s -o $null -w "%{http_code}\n" http://localhost:8080/api/v1/ping \
    -H "Authorization: Bearer $TOKEN"
}
# 101-й запрос вернёт 429
curl -i http://localhost:8080/api/v1/ping \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ при превышении:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
{"error":"rate limit exceeded"}
```

При недоступности Redis запросы пропускаются (fail open).

### 7. Circuit Breaker

Приглашение пользователя обёрнуто в Circuit Breaker (gobreaker). Параметры: MaxRequests=2, Interval=10s, Timeout=30s, порог размыкания ≥3 запроса, ≥60% ошибок.

```bash
# Нормальная работа
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user3@test.com","role":"member"}'
# 200 OK — email отправлен (в логах app)

# Проверка размыкания: измените EMAIL_SERVICE_URL на нерабочий адрес
# EMAIL_SERVICE_URL=http://localhost:9999 в .env, перезапустите
# После 3+ неудач CB разомкнётся — пользователь добавится в БД,
# а в логах появится WARN "circuit breaker open, email not sent"
```

### 8. Prometheus метрики

```bash
curl http://localhost:8080/metrics | grep teamtasks
```

В выводе присутствуют:
- `teamtasks_http_requests_total{method="GET",path="/health",status="OK"}`
- `teamtasks_http_request_duration_seconds_bucket{...}`
- `teamtasks_db_open_connections`
- `teamtasks_db_in_use_connections`
- `teamtasks_db_idle_connections`

Prometheus UI доступен на http://localhost:9090 (после `make up`).

### 9. Healthcheck

```bash
curl http://localhost:8080/health
```

```json
{"status":"ok","database":"connected","cache":"connected","timestamp":"2026-06-25T00:00:00Z"}
```

При недоступности MySQL — `503 Service Unavailable`, `"database":"disconnected"`.

---

## Полный пример flow

```bash
# 1. Регистрация
curl -s -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"pass123"}'

# 2. Логин
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"pass123"}' | \
  powershell -Command "$input | ConvertFrom-Json | Select -ExpandProperty token")

# 3. Создание команды
curl -s -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Dev Team"}'

# 4. Создание задачи
curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"First task","team_id":1}'

# 5. Просмотр метрик
curl http://localhost:8080/metrics | head -20
```

---