# `E0-09` — Embed the SPA, serve with fallback

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-08`

## Story

As a developer, I want the built frontend embedded in the binary and served with an `index.html` fallback, so that there is exactly one artefact to deploy and deep links survive a refresh.

## Scope

**In:**

- `go:embed` of `web/dist` in `internal/infrastructure/web/static.go`.
- A handler serving static assets, falling back to `index.html`.
- A Makefile or script that builds the frontend before the Go binary.

**Out:**

- Docker packaging → `E0-10`.

## Instructions

1. Route precedence: `/api/*` is matched first and **never** falls through to the SPA. An unknown API path must return a JSON 404, not an HTML page — otherwise a frontend bug surfaces as "the API returned HTML", which is a needlessly confusing hour.
2. Everything else: try the embedded file; if absent, serve `index.html` with 200.
3. Hashed assets under `/assets/` get `Cache-Control: public, max-age=31536000, immutable`. `index.html` gets `no-cache` — otherwise a redeploy leaves guests on a stale bundle pointing at hashed files that no longer exist.
4. Set correct MIME types explicitly; do not rely on extension sniffing.
5. The build script runs `pnpm install --frozen-lockfile && pnpm build` in `web/`, then `go build`. A stale `dist/` embedded into a fresh binary is the one skew this design is supposed to make impossible, so the ordering must be enforced by the script, not by memory.
6. Commit an empty `web/dist/.gitkeep` so `go build` works on a clean checkout before the frontend is built, with a comment explaining why it exists.

## Test plan

- [ ] Integration: `GET /` returns `index.html`.
- [ ] Integration: `GET /rsvp` (a client-side route) returns `index.html` with 200.
- [ ] Integration: `GET /api/unknown` returns a JSON 404 envelope, not HTML.
- [ ] Integration: a hashed asset carries the immutable cache header; `index.html` does not.

## Done when

- [ ] One `go build` produces a binary that serves the whole app.
- [ ] Refreshing the browser on a deep link works.
- [ ] Checkbox ticked in `README.md`.
