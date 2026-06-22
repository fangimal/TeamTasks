DOCKER_COMPOSE ?= docker-compose

.PHONY: build up down logs

build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f app
