include .env

DB_DRIVER=postgres
POSTGRES_DSN=host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=${POSTGRES_SSL_MODE}

export POSTGRES_DSN

migrate-up:
	goose -dir ./migrations $(DB_DRIVER) "$(POSTGRES_DSN)" up

migrate-down:
	goose -dir ./migrations $(DB_DRIVER) "$(POSTGRES_DSN)" down

migrate-status:
	goose -dir ./migrations $(DB_DRIVER) "$(POSTGRES_DSN)" status

migrate-reset:
	goose -dir ./migrations $(DB_DRIVER) "$(POSTGRES_DSN)" reset

lint:
	golangci-lint run ./...

security:
	gosec -exclude-dir=bin/database ./...

test:
	go test -cover -coverprofile=./test.out ./...

coverage:
	go tool cover -func ./test.out

coverage-html:
	go tool cover -html=./test.out

validate: lint security test

run:
	go run cmd/main.go

db:
	docker-compose up -d db

up:
	docker-compose up --build

down:
	docker-compose down

restart:
	docker-compose down
	docker-compose up --build

create-migration:
	@read -p "Enter migration name: " name; \
	goose -dir ./migrations create $$name sql

swagger:
	swag init -g ./cmd/api/v1/main.go -o ./docs

.PHONY: run db up down restart migrate-up migrate-down migrate-status migrate-reset