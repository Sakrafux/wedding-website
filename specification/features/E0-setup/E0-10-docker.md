# `E0-10` — Docker image and Compose

**Epic:** E0 — Project setup · **Layer:** ops · **Depends on:** `E0-09`

## Story

As an admin, I want one small container with two volumes, so that deploying is `docker compose up -d` and all durable state is in places I can back up.

## Scope

**In:**

- Multi-stage `Dockerfile`: node build → go build → distroless/scratch runtime.
- `compose.example.yaml` with the volume mounts and environment. The real `compose.yaml` is gitignored: it describes the server — networks, proxy wiring — not the app.
- Non-root runtime user.

**Out:**

- The actual deployment → `E0-12`.

## Instructions

1. Stage 1 builds the frontend. Stage 2 builds the Go binary with `CGO_ENABLED=0` — pure-Go SQLite is what makes a scratch image possible, so this flag is load-bearing, not decoration.
2. Runtime stage: `gcr.io/distroless/static` or `scratch`. Copy the binary and, if using scratch, CA certificates.
3. Run as a non-root UID. Ensure the volume mount points are writable by it — a container that starts and then cannot write to `DB_PATH` fails in a much less obvious way than one that refuses to start.
4. Volumes: `DB_PATH` → `/data/wedding.db`, `PHOTO_DIR` → `/data/photos`. Mount `/data` as one volume; both live under it, which is what makes the backup story "back up one directory".
5. Publish the port to the reverse proxy's network only, not to `0.0.0.0`.
6. Set `restart: unless-stopped`.
7. Layer ordering: copy `go.mod`/`go.sum` and `package.json`/`pnpm-lock.yaml` and fetch dependencies before copying source, so an edit does not re-download the world.
8. Enable pnpm in the build stage via `corepack enable`, with the version pinned by `packageManager` in `package.json`. Corepack means the image needs no global pnpm install, and the pin means the container and your machine resolve the lockfile identically. Use `pnpm install --frozen-lockfile` so a build fails on a stale lockfile instead of quietly resolving something else.
9. `.dockerignore` covering `node_modules`, `.git`, `*.db`, `.env`. A database file accidentally baked into an image is both a fat image and a leaked secret.

## Test plan

- [x] `docker build` succeeds from a clean checkout.
- [x] Image size is under ~30 MB.
- [x] `docker compose up` starts, migrates, and answers the health endpoint.
- [x] The container runs as non-root — verify the UID.
- [x] Stopping and restarting preserves the database in the volume.
- [x] Starting without a required env var fails fast with the message from `E0-02`.

## Done when

- [x] A fresh machine with the repo and Compose can run the app.
- [x] Checkbox ticked in `README.md`.
