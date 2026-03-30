BINARY     := app
CMD        := ./cmd/main.go
MODULE     := github.com/kvitrvn/go-ygg

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags="-s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint run \
        migrate-up migrate-down migrate-version \
        docker-build docker-up docker-down

## ── Go ───────────────────────────────────────────────────────────────────────

build:
	go build $(LDFLAGS) -o bin/$(BINARY) $(CMD)

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

run: build
	./bin/$(BINARY) serve

## ── Migrations ───────────────────────────────────────────────────────────────

migrate-up: build
	./bin/$(BINARY) migrate up

migrate-down: build
	./bin/$(BINARY) migrate down

migrate-version: build
	./bin/$(BINARY) migrate version

## ── Docker ───────────────────────────────────────────────────────────────────

docker-build:
	docker build -t $(BINARY):latest .

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
