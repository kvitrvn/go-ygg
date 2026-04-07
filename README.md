# go-ygg

<p align="center">
  <img src=".github/assets/logo.png" alt="go-ygg logo" width="240" />
</p>

> Go project template — Hexagonal DDD · Cobra · env · golang-migrate

## Requirements

- Go 1.26+
- Docker & Docker Compose (optional)
- [golangci-lint](https://golangci-lint.run/usage/install/) (optional)

## Quick start

```bash
# 1. Configure the application
export GO_YGG_SERVER_PORT=8080
export GO_YGG_DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"

# 2. Download dependencies
go mod download

# 3. Start the server
make run
# → http://localhost:8080/healthz
```

## Commands

| Command                | Description                     |
| ---------------------- | ------------------------------- |
| `make build`           | Compile binary to `bin/`        |
| `make test`            | Run tests with `-race`          |
| `make lint`            | Run golangci-lint               |
| `make run`             | Build + start `serve`           |
| `make install-hooks`   | Enable versioned Git hooks      |
| `make migrate-up`      | Apply all pending migrations    |
| `make migrate-down`    | Revert 1 migration              |
| `make migrate-version` | Print current migration version |
| `make docker-up`       | docker compose up               |
| `make docker-down`     | docker compose down             |

## Git hooks

Enable the versioned pre-commit hook to run the same lint flow as CI:

```bash
make install-hooks
```

The hook regenerates templ and CSS artifacts, then runs `golangci-lint run ./...`.
It executes inside the `app` service from `docker-compose`, so `make up` must already be running.
After changes to `Dockerfile.dev`, rebuild the service with `docker compose up --build -d app`.

## CLI

```
app serve              # Start the HTTP server
app migrate up         # Apply all pending migrations
app migrate down [N]   # Revert N migrations (default: 1)
app migrate version    # Print current version
app --help             # Full help
```

Global configuration is read from environment variables only.

## Configuration

Via environment variables prefixed with `GO_YGG_`:

```bash
GO_YGG_SERVER_HOST=0.0.0.0
GO_YGG_SERVER_PORT=9090
GO_YGG_DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
GO_YGG_LOG_LEVEL=debug
GO_YGG_LOG_FORMAT=text
```

## Architecture

See [docs/technical/architecture.md](docs/technical/architecture.md).

Hexagonal structure:

```
internal/
├── domain/          # Entities, repository interfaces (no external dependencies)
├── application/     # Use cases (depends on domain only)
├── infrastructure/  # Persistence, config (implements domain ports)
└── interfaces/      # CLI (Cobra) + HTTP handlers (inbound adapters)
```

## Adding a DB driver (PostgreSQL example with pgx)

```bash
go get github.com/jackc/pgx/v5
```

In `internal/infrastructure/persistence/`, inject `*pgxpool.Pool`:

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)
```

DSN format: `postgres://user:pass@localhost:5432/dbname?sslmode=disable`

## License

MIT — see [LICENSE.md](LICENSE.md)
