BINARY     := go-ygg
CMD        := ./cmd/main.go
MODULE     := github.com/kvitrvn/go-ygg

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags="-s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)"

CSS_IN  := assets/css/input.css
CSS_OUT := assets/css/output.css

.PHONY: build test lint run generate css dev \
        migrate-up migrate-down migrate-version \
        docker-build docker-up docker-down docker-logs

## ── Codegen ──────────────────────────────────────────────────────────────────

generate:
	templ generate ./...

css:
	tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

dev:
	air

## ── Go ───────────────────────────────────────────────────────────────────────

build: generate css
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

build:
	docker build -t $(BINARY):latest .

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f app
