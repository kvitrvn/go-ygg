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

.PHONY: build run test lint install-hooks \
        up down logs

install-hooks:
	git config core.hooksPath .githooks

## ── Go ───────────────────────────────────────────────────────────────────────

build:
	go build $(LDFLAGS) -o bin/$(BINARY) $(CMD)

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

run: build
	./bin/$(BINARY) serve

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f app
