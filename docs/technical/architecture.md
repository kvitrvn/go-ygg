# Architecture — Hexagonal (Ports & Adapters)

## Dependency rule

Dependencies always point inward. The domain knows nothing about the outside world.

```
┌─────────────────────────────────────────────────────────┐
│  interfaces/          (inbound adapters)                 │
│  ┌──────────┐  ┌──────────┐                             │
│  │ CLI      │  │ HTTP     │                             │
│  └────┬─────┘  └────┬─────┘                             │
│       │              │                                   │
│  ┌────▼──────────────▼─────────────────────────────┐    │
│  │  application/     (use cases)                    │    │
│  │  orchestration, validation, transactions         │    │
│  │  ┌────────────────────────────────────────────┐  │    │
│  │  │  domain/        (pure business logic)      │  │    │
│  │  │  entities, value objects, business rules   │  │    │
│  │  │  repository interfaces (outbound ports)    │  │    │
│  │  └────────────────────────────────────────────┘  │    │
│  └──────────────────────────────────────────────────┘    │
│                                                           │
│  infrastructure/      (outbound adapters)                │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │ persistence/ │  │ config/      │                     │
│  └──────────────┘  └──────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

## Layers

| Layer            | Role                                                      | Depends on       |
|------------------|-----------------------------------------------------------|------------------|
| `domain/`        | Entities, value objects, repository interfaces, rules     | nothing          |
| `application/`   | Use cases, orchestration, inbound ports                   | `domain/`        |
| `infrastructure/`| Port implementations (DB, config, external APIs)          | `domain/`        |
| `interfaces/`    | Inbound adapters (HTTP handlers, CLI commands)            | `application/`   |

## HTTP request flow

```
HTTP request
  → interfaces/http/handler        (decode request)
  → application/example/usecase    (orchestrate)
  → domain/example                 (enforce business rules)
  → infrastructure/persistence     (persist via port)
  → HTTP response                  (encode response)
```

## Adding a bounded context

1. Create `internal/domain/<context>/` — entity + repository interface
2. Create `internal/application/<context>/` — use case(s)
3. Create `internal/infrastructure/persistence/<context>_repository.go`
4. Create `internal/interfaces/http/handler/<context>_handler.go`
5. Register the route in `interfaces/http/router.go`
