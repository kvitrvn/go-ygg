# go-ygg

<p align="center">
  <img src=".github/assets/logo.png" alt="go-ygg logo" width="240" />
</p>

> Opinionated Go blueprint for modular services and modular monoliths with hexagonal boundaries.

`go-ygg` is not meant to be a finished application. It is a reusable starting point for Go projects that want:

- a clear dependency direction
- a small Cobra CLI entrypoint
- env-only configuration
- HTTP and CLI adapters
- migration plumbing
- templ + Tailwind asset generation
- templUI CLI integration without Node.js/npm
- a Docker-based local workflow

The repository already contains the technical skeleton. The business domain is intentionally incomplete and must be replaced when you bootstrap a real project.

## What This Blueprint Gives You

- A single executable entrypoint in [`cmd/main.go`](cmd/main.go)
- A minimal CLI with `serve`, `version`, and `migrate` commands
- HTTP server bootstrap and graceful shutdown
- A sample inbound HTTP adapter with `/`, `/healthz`, and `/version`
- Environment-driven config loading via `GO_YGG_*`
- Migration commands wired through `golang-migrate`
- A placeholder `example` bounded context to show package boundaries
- Docker Compose for local development with PostgreSQL
- `templ`, `templUI`, Tailwind CSS, `air`, and `golangci-lint` in the dev container
- CI and a Docker-based pre-commit hook

## What You Still Need To Build

- Your real bounded contexts, entities, value objects, and use cases
- Real repository implementations and database wiring
- Request/response DTOs beyond the placeholder handlers
- Authentication, authorization, and business-specific policies
- Observability, deployment, and production hardening specific to your system

The current `internal/domain/example` and `internal/application/example` packages are placeholders. They exist to demonstrate structure, not to model a real product domain.

## Bootstrap A New Project

Initialize a new project from this blueprint before doing domain work:

```bash
./scripts/init.sh my-app
```

The script updates the main template identifiers:

- `go-ygg` -> `my-app`
- `GO_YGG_*` -> `MY_APP_*`

After running it, review at least:

- `go.mod`
- `.env.example`
- `README.md`
- package names and user-facing strings
- CI, Docker, and deployment naming

Then replace the sample `example` context with your real domain.

## Quick Start

### Docker Workflow

This is the easiest way to use the blueprint as-is.

```bash
cp .env.example .env
make up
make logs
```

By default the app is exposed on `http://localhost:8080` and the health endpoint is:

```text
http://localhost:8080/healthz
```

The effective HTTP port comes from `GO_YGG_SERVER_PORT` in `.env`.

### Local Host Workflow

If you want to run the app directly on your machine instead of through Docker:

```bash
cp .env.example .env
go mod download
make run
```

This route assumes the generated assets already exist in the workspace. In the intended dev workflow, they are produced by the Docker `app` service through `air`.

Both `make run` and `go run ./cmd/main.go serve` apply pending database migrations before the HTTP server starts, then log whether migrations were applied and which version is active.

## Project Layout

```text
cmd/
  main.go                     # single executable entrypoint

internal/
  domain/                     # entities, value objects, domain contracts
  application/                # use cases and orchestration
  infrastructure/             # config and outbound adapters
  interfaces/                 # inbound adapters: CLI and HTTP
  version/                    # build metadata exposed by the binary

migrations/                   # SQL migrations used by golang-migrate
docs/technical/               # supporting technical documentation
scripts/init.sh               # project bootstrap/rename script
```

The intended dependency rule is simple: adapters and infrastructure point inward; the domain does not depend on HTTP, SQL, config, or framework code.

For the architectural rationale, see [docs/technical/architecture.md](docs/technical/architecture.md).

## Architecture Notes

This blueprint uses a pragmatic hexagonal layout:

- `internal/domain` holds domain types and outbound contracts
- `internal/application` orchestrates use cases
- `internal/interfaces` contains inbound adapters such as Cobra and HTTP
- `internal/infrastructure` contains technical implementations such as config and persistence

The repository is intentionally small. It does not try to demonstrate every DDD pattern up front. Add aggregates, value objects, and ports where they protect real business invariants, not because the architecture diagram says so.

## Development Workflow

### Codegen And Assets

The UI layer uses `templ`, Tailwind CSS, and templUI.

templUI follows the CLI workflow and keeps copied component source inside the HTTP adapter layer:

- `internal/interfaces/http/templui/components`
- `internal/interfaces/http/templui/utils`
- `assets/js/templui`

This preserves the architectural rule that UI concerns stay in `internal/interfaces/http/` and avoids any `nodejs` or `npm` dependency.

In development, the Docker `app` service runs `air`, and `air` regenerates templ output and CSS before rebuilding the binary.

In production, the image build performs the same work in the Dockerfile before compiling the final binary.

### Git Hooks

Enable the versioned pre-commit hook with:

```bash
make install-hooks
```

The hook runs inside the `app` Docker Compose service and expects that service to already be running. It regenerates templ files, rebuilds CSS, and runs `golangci-lint`.

### CI

GitHub Actions currently runs four checks:

- code generation
- lint
- test
- build

Generated templ and CSS artifacts are produced first and reused by later jobs.

## Commands

### CLI

Use the binary directly or via `go run`:

```bash
go run ./cmd/main.go serve
go run ./cmd/main.go version
```

These direct CLI commands run on the host machine. `serve` applies migrations automatically before listening. The standalone `migrate` workflow remains Docker-only in the documented setup.

### Make Targets

The useful `make` targets are:

```bash
make build
make run
make up
make down
make logs
make install-hooks
make test
make lint
```

## Configuration

The application reads configuration from environment variables only.

Default values are documented in `.env.example`:

```bash
GO_YGG_SERVER_HOST=0.0.0.0
GO_YGG_SERVER_PORT=8080
GO_YGG_DATABASE_DSN=postgres://app:secret@db:5432/app?sslmode=disable
GO_YGG_LOG_LEVEL=info
GO_YGG_LOG_FORMAT=json
POSTGRES_USER=app
POSTGRES_PASSWORD=secret
POSTGRES_DB=app
POSTGRES_PORT=5432
```

When you initialize a new project with `scripts/init.sh`, these variable names are renamed to match your project name.

## Database And Migrations

This blueprint wires migration commands through `golang-migrate` and ships with a PostgreSQL-oriented local setup.

When the server starts through `serve`, pending migrations are applied automatically before the HTTP listener is opened. Startup logs indicate whether migrations were applied and which version is now active.

For the normal workflow, you do not need a dedicated make target for migrations.

The expected path is:

```bash
make up
# or
make run
```

If you need to run the migration CLI manually inside Docker, use `docker compose exec -T app sh -lc 'go run ./cmd/main.go migrate ...'`.

The runtime repository implementation in [`internal/infrastructure/persistence/example_repository.go`](internal/infrastructure/persistence/example_repository.go) is still a stub. You are expected to inject your real DB driver and replace the placeholder methods.

More details: [docs/technical/database.md](docs/technical/database.md)

## Extending The Blueprint

When adding a real bounded context:

1. Create domain types and contracts in `internal/domain/<context>`
2. Add use cases in `internal/application/<context>`
3. Implement outbound adapters in `internal/infrastructure/...`
4. Add inbound adapters in `internal/interfaces/...`
5. Register HTTP routes or CLI commands
6. Add focused tests around domain rules and adapter boundaries

Do not keep the sample `example` naming longer than necessary. Rename it early so the ubiquitous language of the codebase matches your actual domain.

## Supporting Docs

- [docs/technical/architecture.md](docs/technical/architecture.md)
- [docs/technical/api.md](docs/technical/api.md)
- [docs/technical/database.md](docs/technical/database.md)

## License

MIT. See [LICENSE.md](LICENSE.md).
