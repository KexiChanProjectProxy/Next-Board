.PHONY: help build run test docker-up docker-down migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make docker-up    - Start Docker Compose services"
	@echo "  make docker-down  - Stop Docker Compose services"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "mysql://xboard:xboard_password@tcp(localhost:3306)/xboard_go" up

migrate-down:
	migrate -path migrations -database "mysql://xboard:xboard_password@tcp(localhost:3306)/xboard_go" down
