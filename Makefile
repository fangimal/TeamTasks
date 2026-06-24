DOCKER_COMPOSE ?= docker-compose
LOCAL_DATABASE_HOST ?= 127.0.0.1

ifeq ($(OS),Windows_NT)
LOCAL_RUN = powershell -NoProfile -ExecutionPolicy Bypass -Command '$$env:DATABASE_HOST="$(LOCAL_DATABASE_HOST)"; go run ./cmd/app'
LOCAL_DEBUG = powershell -NoProfile -ExecutionPolicy Bypass -Command 'if (-not (Get-Command dlv -ErrorAction SilentlyContinue)) { Write-Error "Delve is not installed. Run: go install github.com/go-delve/delve/cmd/dlv@latest"; exit 1 }; $$env:DATABASE_HOST="$(LOCAL_DATABASE_HOST)"; dlv debug ./cmd/app --headless --listen=:40000 --api-version=2 --accept-multiclient'
else
LOCAL_RUN = DATABASE_HOST=$(LOCAL_DATABASE_HOST) go run ./cmd/app
LOCAL_DEBUG = command -v dlv >/dev/null 2>&1 || { echo "Delve is not installed. Run: go install github.com/go-delve/delve/cmd/dlv@latest"; exit 1; }; DATABASE_HOST=$(LOCAL_DATABASE_HOST) dlv debug ./cmd/app --headless --listen=:40000 --api-version=2 --accept-multiclient
endif

.PHONY: build up down logs infra-up local-start local-stop run-local local-debug migrate-up migrate-down test test-unit test-integration test-coverage

build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f app

infra-up:
	$(DOCKER_COMPOSE) up -d mysql redis

local-start: infra-up
	$(LOCAL_RUN)

local-stop:
	$(DOCKER_COMPOSE) stop mysql redis

run-local:
	$(LOCAL_RUN)

local-debug: infra-up
	$(LOCAL_DEBUG)

migrate-up:
	$(DOCKER_COMPOSE) build app
	$(DOCKER_COMPOSE) run --rm --entrypoint /app/teamtasks-migrate app up

migrate-down:
	$(DOCKER_COMPOSE) build app
	$(DOCKER_COMPOSE) run --rm --entrypoint /app/teamtasks-migrate app down

test-unit:
	go test -short -count=1 ./internal/usecase/...

test-integration:
	go test -count=1 -run Integration ./internal/repository/mysql/...

test:
	go test -count=1 ./...

test-coverage:
	go test -short -coverprofile coverage.out -count=1 ./...
	go tool cover -func coverage.out
