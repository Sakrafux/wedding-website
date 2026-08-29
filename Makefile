# Command list for local development. The deployed container gets its environment
# from Compose and never reads .env — the dotenv handling here exists for the dev
# shell alone, which is why nothing in the app parses .env itself.

BINARY := wedding
WEB_DIR := web

# Image registry on the personal server. Plain HTTP, so the pushing and the pulling
# daemon both need it listed under "insecure-registries" — configured on the hosts,
# not here (see E0-12).
REGISTRY := 10.0.0.45:5000
IMAGE := $(REGISTRY)/wedding
TAG ?= latest

# gofmt takes directories, not packages, so the package list is turned back into
# paths. Plain `gofmt -l .` would walk web/node_modules.
GO_DIRS = $(shell go list -f '{{.Dir}}' ./...)

.PHONY: all build build-web run preview test fmt lint clean docker-build docker-push

all: build

## build — frontend first, then the binary that embeds it.
# The ordering is enforced here rather than remembered: a stale web/dist baked into
# a fresh binary is the one frontend/backend skew this whole design exists to rule out.
build: build-web
	go build -o $(BINARY) ./cmd/wedding

## build-web — install with the lockfile and build the bundle into web/dist.
build-web:
	cd $(WEB_DIR) && pnpm install --frozen-lockfile && pnpm build

## run — the Go process against the local .env. Serves whatever was last built into
# web/dist; for frontend work use `pnpm dev` in web/ and open 5173, which proxies
# /api here.
run:
	set -a; . ./.env; set +a; go run ./cmd/wedding

## preview — execute the built binary
preview:
	set -a; . ./.env; set +a; ./$(BINARY)

## test — the full gate: tests, formatting, vet.
test:
	go test ./...
	@test -z "$$(gofmt -l $(GO_DIRS))" || { echo "gofmt needed:"; gofmt -l $(GO_DIRS); exit 1; }
	go vet ./...

fmt:
	gofmt -w $(GO_DIRS)
	cd $(WEB_DIR) && pnpm format

lint:
	cd $(WEB_DIR) && pnpm lint && pnpm typecheck

## docker-build — the deployable image. The frontend is built inside the image, so
# this ignores whatever is currently in web/dist and never depends on `make build`.
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-push — publish to the server's registry. Deployment itself is E0-12.
docker-push: docker-build
	docker push $(IMAGE):$(TAG)

clean:
	rm -f $(BINARY)
	# Keeps .gitkeep: without it the next `go build` has no directory to embed.
	find $(WEB_DIR)/dist -mindepth 1 ! -name .gitkeep -delete
