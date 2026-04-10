# HTTP API

Base URL: `http://localhost:8080`

## Routes

### Public pages

#### `GET /`

Landing page.

### `GET /healthz`

Health check — used by Kubernetes/Docker probes.

**Response `200 OK`**
```json
{ "status": "ok" }
```

### `GET /version`

Build metadata endpoint.

### `GET /signup`

Render the account creation page.

### `POST /signup`

Create a local account, create the personal tenant named from the unique username, create the owner membership, then open a session cookie.

### `GET /login`

Render the sign-in page.

### `POST /login`

Authenticate a local account with email or username and open a session cookie.

### `GET /invitations/{token}`

Render the invitation acceptance page.

### `POST /invitations/{token}/accept`

Accept an invitation.

- if the visitor already has a matching signed-in account, the tenant membership is attached to that account
- otherwise the endpoint signs in an existing account with matching email/password or creates a new account with a personal tenant first

### Authenticated pages

#### `POST /logout`

Delete the current session and clear the auth cookie.

#### `GET /app`

Authenticated workspace home for the active tenant.

#### `GET /app/organizations/new`

Render the organization creation page.

#### `POST /app/organizations`

Create a collaborative organization, attach the current user as `owner`, and switch the active tenant to it.

#### `POST /app/tenants/switch`

Switch the active tenant for the current session.

#### `GET /app/members`

List members of the active organization.

#### `GET /app/invitations/new`

Render the invitation creation page for the active organization.

Requires role `owner` or `admin`.

#### `POST /app/invitations`

Generate a new invitation link for the active organization.

Requires role `owner` or `admin`.

## Conventions

- HTML pages use `Content-Type: text/html; charset=utf-8`
- JSON endpoints currently remain limited to `/healthz` and `/version`
- Routes use the Go 1.22 `METHOD /path/{param}` syntax
- Session auth uses an opaque `HttpOnly` cookie
- Unsafe methods rely on same-origin checks using `Origin` or `Referer`

## Adding a route

1. Create the handler in the package that owns the feature, currently `internal/interfaces/http/web/` for auth-backed pages
2. Register it in `internal/interfaces/http/router.go`:
   ```go
   mux.HandleFunc("GET /examples/{id}", handler.GetExample)
   ```
3. Document it here.
