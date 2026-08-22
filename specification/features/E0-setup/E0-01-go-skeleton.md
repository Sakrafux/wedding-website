# `E0-01` — Go module, package layout, router

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** nothing

## Story

As a developer, I want the module and package skeleton in place so that every later story has an obvious home and the layering rules are enforced from the first commit rather than retrofitted.

## Scope

**In:**

- `go.mod`, Go 1.23+.
- The directory tree from [04-architecture](../../04-architecture.md), with a `doc.go` or a placeholder in each package so the layout is visible in an empty repo.
- chi router mounted in `cmd/wedding`, `httplog` middleware, `GET /api/health` returning `{"status":"ok"}`.
- Graceful shutdown on SIGINT/SIGTERM with a read/write/idle timeout on the server.

**Out:**

- Config parsing → `E0-02`.
- Database → `E0-03`.
- Static file serving → `E0-09`.

## Instructions

1. Create the tree: `cmd/wedding/`, `internal/domain/`, `internal/application/`, `internal/infrastructure/{web,web/dto,web/middleware,persistence,configuration,security,photo}`, `tests/integration/`, `web/`.
2. Add only these direct dependencies for now: `go-chi/chi/v5`, `go-chi/httplog/v2`. Resist adding the rest until the story that needs them.
3. `cmd/wedding/main.go` wires: logger → router → server. Nothing else lives in `main`.
4. Set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` explicitly. Go's zero values mean "no timeout", which is how a small server gets held open by a slow client.
5. Health endpoint is unauthenticated and stays that way — it is what the reverse proxy and any restart check will poll.

## Contract

```http
GET /api/health
```

Response `200`:

```json
{ "status": "ok" }
```

## Test plan

- [ ] Integration: `GET /api/health` returns 200 and the expected body.
- [ ] Unknown `/api/*` path returns 404 as JSON, not as HTML.
- [ ] `go vet ./...` clean.

## Done when

- [ ] `go run ./cmd/wedding` serves the health endpoint locally.
- [ ] The package tree exists and matches the architecture document.
- [ ] Checkbox ticked in `README.md`.
