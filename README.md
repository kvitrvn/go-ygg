# go-ygg

> Go project template — Hexagonal DDD · Cobra · Viper · golang-migrate

## Requirements

- Go 1.26+
- Docker & Docker Compose (optional)
- [golangci-lint](https://golangci-lint.run/usage/install/) (optional)

## Quick start

```bash
# 1. Copy and edit the configuration
cp config.example.yaml config.yaml

# 2. Download dependencies
go mod download

# 3. Start the server
make run
# → http://localhost:8080/healthz
```

## Commands

| Command               | Description                          |
|-----------------------|--------------------------------------|
| `make build`          | Compile binary to `bin/`             |
| `make test`           | Run tests with `-race`               |
| `make lint`           | Run golangci-lint                    |
| `make run`            | Build + start `serve`                |
| `make migrate-up`     | Apply all pending migrations         |
| `make migrate-down`   | Revert 1 migration                   |
| `make migrate-version`| Print current migration version      |
| `make docker-up`      | docker compose up                    |
| `make docker-down`    | docker compose down                  |

## CLI

```
app serve              # Start the HTTP server
app migrate up         # Apply all pending migrations
app migrate down [N]   # Revert N migrations (default: 1)
app migrate version    # Print current version
app --help             # Full help
```

Global flags: `--config <path>`, `--log-level debug|info|warn|error`

## Configuration

Via `config.yaml` (copied from `config.example.yaml`) or environment variables prefixed with `APP_`:

```bash
APP_SERVER_PORT=9090
APP_DATABASE_DSN="postgres://..."
APP_LOG_LEVEL=debug
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
go get github.com/golang-migrate/migrate/v4/database/pgx/v5
```

In `internal/infrastructure/persistence/`, add the blank imports and inject `*pgxpool.Pool`:

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)
```

DSN format: `pgx5://user:pass@localhost:5432/dbname?sslmode=disable`

## License

MIT — see [LICENSE.md](LICENSE.md)
