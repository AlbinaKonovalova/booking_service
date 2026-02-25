.PHONY: build run migrate-up migrate-down migrate-create lint test clean deps

# Переменные
BINARY_NAME=server
DATABASE_URL?=postgres://postgres:postgres@localhost:5432/booking_service?sslmode=disable
MIGRATIONS_DIR=migrations

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Миграции goose
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

# Запуск сервера
run:
	go run cmd/server/main.go

# Сборка
build:
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

# Линтинг
lint:
	golangci-lint run ./...

# Тесты
test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Очистка
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

