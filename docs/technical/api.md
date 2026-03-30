# HTTP API

Base URL: `http://localhost:8080`

## Routes

### `GET /healthz`

Health check — used by Kubernetes/Docker probes.

**Response `200 OK`**
```json
{ "status": "ok" }
```

---

## Conventions

- All responses use `Content-Type: application/json`
- Errors follow the structure:
  ```json
  { "error": "human-readable message" }
  ```
- Routes use the Go 1.22 `METHOD /path/{param}` syntax

## Adding a route

1. Create the handler in `internal/interfaces/http/handler/`
2. Register it in `internal/interfaces/http/router.go`:
   ```go
   mux.HandleFunc("GET /examples/{id}", handler.GetExample)
   ```
3. Document it here.
