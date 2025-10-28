include .env

DB_DRIVER=postgres
POSTGRES_DSN=host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=${POSTGRES_SSL_MODE}

export POSTGRES_DSN

TEST_PKGS:=$(shell go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
COVERAGE_UNIT:=coverage-unit.out
COVERAGE_INT:=coverage-integration.out

define run_go_test
	go test $(1) -count=1 $(TEST_PKGS)
endef

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

generate:
	go generate ./...

install:
	go mod download
	go install github.com/vektra/mockery/v2@latest

clean:
	rm -rf coverage-*.out internal/service/mocks

coverage: generate
	rm -f $(COVERAGE_UNIT)
	$(call run_go_test,-coverprofile=$(COVERAGE_UNIT))
	go tool cover -func=$(COVERAGE_UNIT)

coverage-html: coverage
	go tool cover -html=$(COVERAGE_UNIT)

coverage-integration: generate
	rm -f $(COVERAGE_INT)
	$(call run_go_test,-tags=integration -coverprofile=$(COVERAGE_INT))
	go tool cover -func=$(COVERAGE_INT)

coverage-integration-html: coverage-integration
	go tool cover -html=$(COVERAGE_INT)

validate: lint security test

run: swagger
	go run cmd/main.go

test: generate
	$(call run_go_test,)

test-integration: generate
	$(call run_go_test,-tags=integration)

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
	swag init -g ./cmd/main.go -o ./docs

.PHONY: run db up down restart migrate-up migrate-down migrate-status migrate-reset swagger test test-integration lint security generate coverage coverage-html coverage-integration coverage-integration-html validate
