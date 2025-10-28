-include .env

DB_DRIVER = postgres
POSTGRES_DSN = host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=${POSTGRES_SSL_MODE}

export POSTGRES_DSN
export GO ?= go
export GOPROXY ?= https://proxy.golang.org,direct
export GOBIN ?= $(shell $(GO) env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

HOST_GOOS := $(shell $(GO) env GOOS)
HOST_GOARCH := $(shell $(GO) env GOARCH)
APP_GOOS ?= linux
APP_GOARCH ?= $(HOST_GOARCH)

BIN_DIR := bin
BIN_HOST := $(BIN_DIR)/server.$(HOST_GOOS)-$(HOST_GOARCH)
BIN_APP := $(BIN_DIR)/server.$(APP_GOOS)-$(APP_GOARCH)
BIN_MAIN := $(BIN_DIR)/server
BUILD_MAIN := ./cmd/main.go

COVERAGE_DIR := coverage
COVERAGE_UNIT := $(COVERAGE_DIR)/unit.out
COVERAGE_INT := $(COVERAGE_DIR)/integration.out

TEST_PKGS := $(shell $(GO) list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
TEST_ARGS ?=
DOCKER_COMPOSE ?= docker-compose
GOSEC_BIN := $(GOBIN)/gosec
SWAG_BIN := $(GOBIN)/swag
GOOSE_BIN := $(GOBIN)/goose
STATICCHECK_BIN := $(GOBIN)/staticcheck
STATICCHECK_DIRS := $(shell $(GO) list -f '{{.Dir}}' ./...)

.DEFAULT_GOAL := help

help:
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z0-9_.%-]+:.*##/ {printf "  %-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: 
	$(GO) mod tidy 
	$(GO) mod download
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) install github.com/vektra/mockery/v2@latest
	$(GO) install github.com/pressly/goose/v3/cmd/goose@v3.26.0
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest

generate: | install
	$(GO) generate ./...

lint: install generate
	$(GO) vet ./...
	@for dir in $(STATICCHECK_DIRS); do \
		$(STATICCHECK_BIN) "$$dir" || exit $$?; \
	done

security: install generate
	$(GOSEC_BIN) -exclude-dir=bin/database ./...

test: install generate 
	$(GO) test -count=1 $(TEST_ARGS) $(TEST_PKGS)

validate: lint security test

test-integration: install generate 
	$(GO) test -count=1 -tags=integration $(TEST_ARGS) $(TEST_PKGS)

coverage: generate 
	@mkdir -p $(COVERAGE_DIR)
	rm -f $(COVERAGE_UNIT)
	$(GO) test -count=1 -coverprofile=$(COVERAGE_UNIT) $(TEST_ARGS) $(TEST_PKGS)
	$(GO) tool cover -func=$(COVERAGE_UNIT)

coverage-html: coverage
	$(GO) tool cover -html=$(COVERAGE_UNIT)

coverage-integration: generate
	@mkdir -p $(COVERAGE_DIR)
	rm -f $(COVERAGE_INT)
	$(GO) test -count=1 -tags=integration -coverprofile=$(COVERAGE_INT) $(TEST_ARGS) $(TEST_PKGS)
	$(GO) tool cover -func=$(COVERAGE_INT)

coverage-integration-html: coverage-integration 
	$(GO) tool cover -html=$(COVERAGE_INT)

build: validate
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) build -o $(BIN_HOST) $(BUILD_MAIN)
	@cp $(BIN_HOST) $(BIN_MAIN)

build-container: validate
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=$(APP_GOOS) GOARCH=$(APP_GOARCH) $(GO) build -o $(BIN_APP) $(BUILD_MAIN)
	@cp $(BIN_APP) $(BIN_MAIN)

app: build-container
	$(DOCKER_COMPOSE) up -d --build app

run: generate
	$(GO) run $(BUILD_MAIN)

db:
	$(DOCKER_COMPOSE) up -d db

up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

restart: 
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up --build

migrate-%: install 
	$(GOOSE_BIN) -dir ./migrations $(DB_DRIVER) "$(POSTGRES_DSN)" $*

create-migration: install 
	@read -p "Enter migration name: " name; \
	$(GOOSE_BIN) -dir ./migrations create $$name sql

swagger: install
	$(SWAG_BIN) init -g ./cmd/main.go -o ./docs

clean: 
	rm -rf $(BIN_DIR) $(COVERAGE_DIR) internal/service/mocks

PHONY_TARGETS := \
	help install generate lint security test check test-integration coverage coverage-html \
	coverage-integration coverage-integration-html build build-container app run \
	validate db up down restart migrate-% create-migration swagger clean

.PHONY: $(PHONY_TARGETS)
