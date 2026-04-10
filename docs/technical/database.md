# Database & Migrations

## Current State

This blueprint is currently wired for PostgreSQL migrations through [`golang-migrate`](https://github.com/golang-migrate/migrate) with:

- migration source: `file://migrations`
- migration database driver: `github.com/golang-migrate/migrate/v4/database/postgres`
- default local database: the PostgreSQL service from `docker-compose.yml`

The project now includes a first real PostgreSQL-backed persistence slice for the `iam` bounded context in [`internal/infrastructure/persistence/postgres.go`](../../internal/infrastructure/persistence/postgres.go). The old example repository remains a placeholder for the sample bounded context only.

## What The `serve` Command Does

Starting the HTTP server also applies pending migrations before the listener is opened.

The flow implemented in [`internal/interfaces/cli/serve.go`](../../internal/interfaces/cli/serve.go) is:

1. load config
2. create a `golang-migrate` client
3. run `migrate up`
4. log the resulting version and dirty state
5. start the HTTP server

If there is nothing to apply, startup logs report that the database is already up to date.

## Migration CLI

The dedicated migration commands are implemented in [`internal/interfaces/cli/migrate.go`](../../internal/interfaces/cli/migrate.go).

Available commands:

```bash
go run ./cmd/main.go migrate up
go run ./cmd/main.go migrate down 1
go run ./cmd/main.go migrate version
```

In the normal Docker workflow, run them inside the `app` service:

```bash
docker compose exec -T app sh -lc 'go run ./cmd/main.go migrate up'
docker compose exec -T app sh -lc 'go run ./cmd/main.go migrate down 1'
docker compose exec -T app sh -lc 'go run ./cmd/main.go migrate version'
```

The `version` command prints:

- `version: none, dirty: false` when no migration has been applied yet
- `version: <N>, dirty: <bool>` otherwise

## Standard Workflow

For the usual development flow, you do not need to call the migration CLI manually.

Use:

```bash
make up
# or
make run
```

Both paths end up starting `serve`, which applies pending migrations automatically.

## Migration Files

The project currently ships one initial migration:

```text
migrations/
  000001_init.up.sql
  000001_init.down.sql
```

That initial migration creates the current auth schema:

- `users`
- `tenants`
- `memberships`
- `invitations`
- `sessions`

The current schema models:

- a global `users.username` unique identifier in addition to unique email
- a `tenants.is_personal` flag to distinguish personal tenants from collaborative organizations
- memberships that link users to both their personal tenant and any organizations they join
- invitations that apply only to collaborative organizations

Follow the existing convention for new migrations:

- 6-digit sequential numbering
- one `.up.sql` and one matching `.down.sql`
- one coherent schema change per migration

Example:

```text
migrations/
  000001_init.up.sql
  000001_init.down.sql
  000002_add_users.up.sql
  000002_add_users.down.sql
```

## Configuration

Database connectivity is configured through `GO_YGG_DATABASE_DSN`.

Example:

```bash
export GO_YGG_DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

In the Docker Compose setup, the app container defaults to the PostgreSQL service defined in `docker-compose.yml`.

Auth-related runtime config also depends on:

```bash
GO_YGG_APP_BASE_URL=http://localhost:8080
GO_YGG_AUTH_COOKIE_NAME=go_ygg_session
GO_YGG_AUTH_COOKIE_SECURE=false
GO_YGG_AUTH_SESSION_TTL=168h
GO_YGG_AUTH_INVITATION_TTL=168h
```

## If You Change Database Engine Later

The blueprint is not database-agnostic out of the box today.

If you move away from PostgreSQL, you will need to update at least:

- the blank import in [`internal/interfaces/cli/migrate.go`](../../internal/interfaces/cli/migrate.go)
- the runtime SQL driver opened in [`internal/infrastructure/persistence/postgres.go`](../../internal/infrastructure/persistence/postgres.go)
- the DSN format in your environment
- the local container setup in `docker-compose.yml`
- your real persistence implementation once you replace the example repository
