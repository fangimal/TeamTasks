# TeamTasks

REST API сервис для управления задачами в командах. Текущий этап: инфраструктурный каркас приложения.

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

## Проверка работоспособности

Healthcheck:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{
  "status": "ok",
  "timestamp": "2026-06-21T00:00:00Z"
}
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
- `REDIS_HOST`
- `REDIS_PORT`
- `LOGGER_LEVEL`
- `LOGGER_FORMAT`

Чтобы изменить порт приложения локально, нужно поменять `SERVER_PORT` в `.env` и перезапустите контейнеры через `make up`.
