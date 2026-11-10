GO ?= go
TEST_FLAGS ?= -v -race -parallel 5 -shuffle=on
COVER_FLAGS ?= -coverprofile=cover.out -covermode=atomic -coverpkg=./...
BINARY_NAME ?= bin/http-server
LINT_VERSION := v2.15.0
MOCKERY_VERSION := v2.53.8

.PHONY: help docker-up docker-build docker-clean test lint deps build clean mock
.DEFAULT_GOAL := help

docker-up:
	docker compose up -d --build

docker-build:
	docker compose build

docker-clean:
	docker compose down

test:
	$(GO) test $(TEST_FLAGS) $(COVER_FLAGS) ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION) run ./...

deps:
	$(GO) mod download
	$(GO) mod verify

build: deps
	$(GO) build -o $(BINARY_NAME) ./cmd/http-server

clean:
	rm -rf bin/
	rm -f cover.out

mock:
	$(GO) run github.com/vektra/mockery/v2@$(MOCKERY_VERSION) --config config/.mockery.yaml
