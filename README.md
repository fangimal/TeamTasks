# TeamTasks

REST API сервис для управления задачами в командах. Текущий этап: отказоустойчивость (Rate Limiting, Circuit Breaker).

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
- [x] создание, просмотр, обновление задач;
- [x] назначение ответственного за задачу;
- [x] статусы задач: `todo`, `in_progress`, `done`;
- [x] фильтрация и пагинация списка задач;
- [x] история изменений задач (аудит);
- [x] комментарии к задачам;
- [x] кэширование списков задач через Redis;
- [x] сложные SQL-запросы для аналитики и выборок по командам;
- [x] rate limiting (100 запр/мин на пользователя, Redis + INCR/EXPIRE);
- [x] circuit breaker для внешнего email-сервиса (gobreaker);
- [ ] Prometheus-метрики через `/metrics`.

Текущее состояние проекта: реализована инфраструктура приложения, подключение к MySQL и Redis, миграции схемы БД, базовый слой Repository, проверка подключения к базе данных и Redis в `/health`, регистрация, логин, JWT middleware, создание команд, получение списка команд, приглашение участников с проверкой ролей, полный CRUD для задач с фильтрацией, пагинацией и проверкой прав доступа (RBAC), аудит изменений задач с транзакционной записью в `task_history`, комментарии к задачам с проверкой членства в команде, кэширование списков задач в Redis (TTL 5 минут) с автоматической инвалидацией при создании или обновлении задачи, аналитические эндпоинты со сложными SQL-запросами (статистика команд, топ пользователей с оконными функциями, проверка целостности данных) и оптимизированные индексы, rate limiting (100 запр/мин на пользователя через Redis), circuit breaker для внешнего email-сервиса (gobreaker) с graceful degradation.

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
  "cache": "connected",
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

### Задачи

Создание задачи (автор должен быть участником команды):

```bash
TOKEN="<token-from-login>"
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Implement auth","description":"Add JWT auth","assignee_id":1,"team_id":1}'
```

Ожидаемый ответ: `201 Created` с данными задачи. Поля `status` и `description` опциональны (по умолчанию `todo` и пустая строка).

Попытка создать задачу в команде, где пользователь не состоит:

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Hack","assignee_id":1,"team_id":999}'
```

Ожидаемый ответ: `403 Forbidden` с сообщением "you are not a member of this team".

Попытка назначить исполнителя, не состоящего в команде:

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Bad assignee","assignee_id":999,"team_id":1}'
```

Ожидаемый ответ: `400 Bad Request` с сообщением "assignee is not a member of this team".

Получение списка задач команды с фильтрацией и пагинацией:

```bash
curl -i "http://localhost:8080/api/v1/tasks?team_id=1&status=todo&limit=5&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ: `200 OK` с JSON вида:
```json
{
  "data": [ { "id": 1, "title": "Implement auth", ... } ],
  "total": 1,
  "limit": 5,
  "offset": 0
}
```

Фильтр по исполнителю:

```bash
curl -i "http://localhost:8080/api/v1/tasks?team_id=1&assignee_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

Получение задачи по ID:

```bash
curl -i http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

Пользователь не из команды пытается получить задачу — `403 Forbidden`.

Обновление задачи:

```bash
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Updated title","status":"in_progress"}'
```

Ожидаемый ответ: `200 OK` с обновлёнными данными задачи.

Проверка RBAC — участник с ролью `member` пытается обновить чужую задачу (не является assignee и не автор):

```bash
TOKEN_MEMBER="<token-of-member-user-not-owner-of-task>"
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_MEMBER" \
  -d '{"title":"Hacked"}'
```

Ожидаемый ответ: `403 Forbidden` с сообщением "insufficient permissions".

### История изменений задач

При каждом обновлении задачи (`PUT /api/v1/tasks/{id}`) автоматически создаётся запись в таблице `task_history`. Обновление задачи и запись истории выполняются в одной транзакции: если запись в историю не удалась, обновление задачи откатывается.

Если отправлен `PUT` с данными, идентичными текущим, запись в историю **не создаётся**.

Получение истории задачи (участник команды):

```bash
curl -i "http://localhost:8080/api/v1/tasks/1/history?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK`:

```json
{
  "data": [
    {
      "id": 1,
      "task_id": 1,
      "changed_by": 1,
      "changed_at": "2026-06-22T19:55:12Z",
      "old_value": {"status": "todo"},
      "new_value": {"status": "in_progress"}
    }
  ],
  "limit": 10,
  "offset": 0
}
```

Пользователь, не состоящий в команде задачи, получает `403 Forbidden`.

### Комментарии к задачам

Создание комментария (участник команды):

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"text":"Great job!"}'
```

Ожидаемый ответ: `201 Created` с данными комментария.

Пользователь, не состоящий в команде задачи, получает `403 Forbidden`:

```bash
TOKEN_OTHER="<token-of-user-not-in-team>"
curl -i -X POST http://localhost:8080/api/v1/tasks/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_OTHER" \
  -d '{"text":"Hack"}'
```

Получение комментариев с пагинацией:

```bash
curl -i "http://localhost:8080/api/v1/tasks/1/comments?limit=5&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK`:
```json
{
  "data": [
    {
      "id": 1,
      "task_id": 1,
      "user_id": 1,
      "text": "Great job!",
      "created_at": "2026-06-23T12:00:00Z",
      "updated_at": "2026-06-23T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 5,
  "offset": 0
}
```

Комментарий к несуществующей задаче возвращает `404 Not Found`:

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks/9999/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"text":"test"}'
```

Пустой текст комментария возвращает `400 Bad Request`:

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"text":""}'
```

### Кэширование списков задач

Список задач команды (`GET /api/v1/tasks?team_id=1`) кэшируется в Redis с TTL 5 минут.

Ключ кэша строится на основе параметров фильтра (`team_id`, `status`, `assignee_id`) и пагинации (`limit`, `offset`). Первый запрос после сохранения данных выполняется из Redis (1-2 мс), последующие — из MySQL (10-50 мс).

**Инвалидация кэша:**
- При создании задачи (`POST /api/v1/tasks`) — кэш команды очищается.
- При обновлении задачи (`PUT /api/v1/tasks/{id}`) — кэш команды очищается.

Проверка кэширования:

```bash
# Первый запрос (Cache Miss — данные из MySQL)
time curl -s "http://localhost:8080/api/v1/tasks?team_id=1" \
  -H "Authorization: Bearer $TOKEN"

# Второй запрос (Cache Hit — данные из Redis)
time curl -s "http://localhost:8080/api/v1/tasks?team_id=1" \
  -H "Authorization: Bearer $TOKEN"

# Проверка инвалидации — обновляем задачу
curl -i -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Updated after cache"}'

# Следующий запрос снова будет Cache Miss
```

Проверка ключей в Redis (через `redis-cli`):

```bash
docker-compose exec redis redis-cli keys 'tasks:team:*'
```

Логи приложения:

```bash
make logs
```

Остановка контейнеров:

```bash
make down
```

### Аналитика и сложные SQL-запросы

Статистика по командам (количество участников, выполненных задач за 7 дней):

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

Топ-3 пользователей по количеству созданных задач в каждой команде (с оконной функцией `DENSE_RANK`):

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

Проверка целостности данных — задачи, у которых `assignee_id` не состоит в `team_members` команды:

```bash
curl -i http://localhost:8080/api/v1/analytics/integrity-check \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ `200 OK` (пустой массив при отсутствии нарушений):
```json
[]
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
- `REDIS_PASSWORD`
- `REDIS_DB`
- `LOGGER_LEVEL`
- `LOGGER_FORMAT`
- `JWT_SECRET`
- `JWT_EXPIRATION`

дополнение:

- `RATE_LIMIT_ENABLED` — включить/выключить лимит (true/false)
- `RATE_LIMIT_REQUESTS_PER_MINUTE` — максимальное число запросов в минуту (100)
- `EMAIL_SERVICE_URL` — URL внешнего email-сервиса
- `EMAIL_SERVICE_TIMEOUT` — таймаут HTTP-запроса к email-сервису (5s)

Чтобы изменить порт приложения локально, нужно поменять `SERVER_PORT` в `.env` и перезапустите контейнеры через `make up`.

## Rate Limiting

На все защищённые эндпоинты (требующие JWT) наложено ограничение: **100 запросов в минуту на одного пользователя**. Счётчик хранится в Redis (ключ `rate_limit:{user_id}`). Для неавторизованных запросов лимит не применяется.

При превышении лимита возвращается `429 Too Many Requests`:

```bash
# Быстрая серия запросов для проверки
for ($i=0; $i -le 100; $i++) {
  curl -s -o $null -w "%{http_code}\n" http://localhost:8080/api/v1/ping \
    -H "Authorization: Bearer $TOKEN"
}
# 101-й запрос вернёт 429
curl -i http://localhost:8080/api/v1/ping \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: application/json

{"error":"rate limit exceeded"}
```

При недоступности Redis запросы пропускаются (fail open).

## Circuit Breaker (Email Service)

Приглашение пользователя в команду (`POST /teams/{id}/invite`) отправляет HTTP-запрос к внешнему email-сервису (по умолчанию `httpbin.org/post`). Вызов обёрнут в `Circuit Breaker` (`gobreaker`).

**Параметры CB:**
- `MaxRequests: 2` (сколько запросов пропускать в полуоткрытом состоянии)
- `Interval: 10s` (интервал сброса счётчиков)
- `Timeout: 30s` (время в разомкнутом состоянии перед переходом в полуоткрытое)
- Порог размыкания: ≥3 запроса, ≥60% ошибок

**Проверка нормальной работы:**

```bash
curl -i -X POST http://localhost:8080/api/v1/teams/1/invite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user@test.com","role":"member"}'
# 200 OK — пользователь добавлен, email отправлен (в логах app)
```

**Проверка размыкания CB:**

Измените `EMAIL_SERVICE_URL` в `.env` на нерабочий адрес (например, `http://localhost:9999`), перезапустите приложение и сделайте 5+ запросов на приглашение. После размыкания CB запросы перестанут пытаться соединиться с сервером (мгновенный ответ), пользователь всё равно будет добавлен в БД, а в логах появится `WARN circuit breaker open, email not sent`.