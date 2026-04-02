# Database & Migrations

## Driver

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) with a `file://migrations` source.

The DB driver is **agnostic**: add the appropriate blank import in `internal/infrastructure/persistence/`.

**Recommended: pgx v5** (actively maintained, native PostgreSQL protocol)

```bash
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4/database/pgx/v5
```

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)
```

Other supported drivers:

| Database | golang-migrate import                                          | DSN scheme  |
|----------|----------------------------------------------------------------|-------------|
| pgx v5   | `github.com/golang-migrate/migrate/v4/database/pgx/v5`        | `pgx5://`   |
| MySQL    | `github.com/golang-migrate/migrate/v4/database/mysql`         | `mysql://`  |
| SQLite3  | `github.com/golang-migrate/migrate/v4/database/sqlite3`       | `sqlite3://`|

## File naming convention

```
migrations/
  000001_init.up.sql
  000001_init.down.sql
  000002_add_users.up.sql
  000002_add_users.down.sql
```

- 6-digit sequential numbering
- Always provide the matching `.down.sql`
- One migration = one coherent, atomic change

## Commands

```bash
make migrate-up            # apply all pending migrations
make migrate-down          # revert 1 migration
./bin/app migrate down 3   # revert 3 migrations
make migrate-version       # print current version
```

## DSN

Configure via `config.yaml` or the `GO_YGG_DATABASE_DSN` environment variable:

```bash
export GO_YGG_DATABASE_DSN="pgx5://user:pass@localhost:5432/dbname?sslmode=disable"
```
