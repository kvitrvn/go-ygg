# Database & Migrations

## Driver

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) with a `file://migrations` source.

The DB driver is **agnostic**: add the appropriate blank import in `internal/infrastructure/persistence/`.

**Recommended runtime driver: pgx v5** (actively maintained, native PostgreSQL protocol)

```bash
go get github.com/jackc/pgx/v5
```

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)
```

Other supported drivers:

| Database | golang-migrate import                                          | DSN scheme  |
|----------|----------------------------------------------------------------|-------------|
| PostgreSQL | `github.com/golang-migrate/migrate/v4/database/postgres`     | `postgres://` |
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

Configure via the `GO_YGG_DATABASE_DSN` environment variable:

```bash
export GO_YGG_DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```
