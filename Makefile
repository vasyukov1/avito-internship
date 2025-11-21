.PHONY: build run up down logs fmt lint test

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
