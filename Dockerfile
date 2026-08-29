# syntax=docker/dockerfile:1

# Three stages: build the bundle, build the binary that embeds it, ship the binary
# alone. The result is one artefact with no runtime toolchain and no shell.

# --- Stage 1: frontend --------------------------------------------------------
FROM node:24-alpine AS web

# Corepack resolves pnpm from the "packageManager" pin in package.json, so the image
# needs no global pnpm install and resolves the lockfile exactly like a dev machine.
# The prompt would otherwise block a non-interactive build.
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable

WORKDIR /build/web

# Manifest and lockfile first: a source edit must not re-download the dependency
# tree. --frozen-lockfile fails the build on a stale lockfile instead of quietly
# resolving something the developer never tested.
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# --- Stage 2: binary ----------------------------------------------------------
FROM golang:1.26-alpine AS api

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
# Only the embed declaration is needed from web/; the bundle itself comes from the
# stage above, which is what keeps a stale local web/dist out of the image.
COPY web/embed.go ./web/embed.go
COPY --from=web /build/web/dist ./web/dist

# CGO_ENABLED=0 is load-bearing, not decoration: modernc.org/sqlite is pure Go, and
# a static binary is what allows the scratch-like runtime stage below.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/wedding ./cmd/wedding

# Staged here because the runtime image has no shell to mkdir with. A named volume
# mounted at /data inherits these directories and their ownership from the image, so
# the non-root user can write on a first start.
RUN mkdir -p /out/data/photos

# --- Stage 3: runtime ---------------------------------------------------------
# distroless/static: no shell, no package manager, CA certificates included.
FROM gcr.io/distroless/static-debian12:nonroot

# nonroot is UID/GID 65532 in this base image.
COPY --from=api --chown=nonroot:nonroot /out/data /data
COPY --from=api /out/wedding /usr/local/bin/wedding

# Defaults matching the volume layout below. Everything else — credentials, proxy
# CIDRs — is required at startup and deliberately has no default.
ENV DB_PATH=/data/wedding.db \
    PHOTO_DIR=/data/photos

EXPOSE 8080
USER nonroot:nonroot

# No HEALTHCHECK: the image has no shell, curl or wget to run one with. Liveness is
# the reverse proxy's job, and GET /api/health is there for it.
ENTRYPOINT ["/usr/local/bin/wedding"]
