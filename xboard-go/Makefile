SHELL := /bin/bash
.PHONY: help build run test migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make clean        - Remove built binaries"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database "mysql://root@tcp(localhost:3306)/xboard" up

migrate-down:
	migrate -path migrations -database "mysql://root@tcp(localhost:3306)/xboard" down

clean:
	rm -rf bin/
