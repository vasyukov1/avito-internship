.PHONY: build run up down logs fmt lint test migrate-up migrate-down migrate-force

include .env
export $(shell sed 's/=.*//' .env)

MIGRATE=migrate
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down

migrate-force:
	$(MIGRATE) -path migrations -database "$(DB_URL)" force 1

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

up:
	docker-compose up --build -d

down:
	docker-compose down

logs:
	docker-compose logs -f app
