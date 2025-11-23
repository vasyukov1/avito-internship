.PHONY: build run up down logs fmt lint test migrate-up migrate-down migrate-force monitoring

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

monitoring:
	docker-compose up -d prometheus grafana

monitoring-logs:
	docker-compose logs -f prometheus grafana

monitoring-stop:
	docker-compose stop prometheus grafana

full-up:
	docker-compose up -d
	@echo "Application: http://localhost:8080"
	@echo "Swagger: http://localhost:8080/swagger/index.html"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"
