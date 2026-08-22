# `E0-02` — Configuration from environment

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-01`

## Story

As a developer, I want configuration read from the environment and validated at startup so that a missing variable stops the container immediately instead of surfacing as a strange bug three weeks later.

## Scope

**In:**

- `internal/infrastructure/configuration`: a `Config` struct and a `Load()` that reads the environment, validates, and returns an error naming **every** missing or invalid variable at once.
- The variable set from [04-architecture](../../04-architecture.md): `PORT`, `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`, `SESSION_COOKIE_SECURE`, `TRUSTED_PROXY_CIDRS`, `LOG_LEVEL`.
- `.env.example` in the repo root, with placeholder values and a comment per variable.

**Out:**

- Anything reading the config beyond `main` — later stories take it as a parameter.

## Instructions

1. Required: `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`. Absent → hard fail, no defaults.
2. Optional with defaults: `PORT` = 8080, `LOG_LEVEL` = info, `SESSION_COOKIE_SECURE` = true, `TRUSTED_PROXY_CIDRS` = empty.
3. `SESSION_COOKIE_SECURE` defaults to **true**. Insecure must be the deliberate act, not the accident.
4. Parse `TRUSTED_PROXY_CIDRS` into `[]netip.Prefix` at load time. A malformed CIDR fails startup — a silently ignored one makes rate limiting bypassable, which is exactly the failure mode nobody notices.
5. Reject an `ADMIN_PASSWORD` shorter than 12 characters. Cheap guard against a placeholder reaching production.
6. `Config` has a `String()`/`LogValue()` that **redacts `ADMIN_PASSWORD`**. Startup logging that dumps the config is normal; leaking the admin password into the host's log collector is not.
7. Collect all validation problems and report them together. Fixing env vars one restart at a time is miserable.

## Test plan

- [ ] Unit: all required vars present → valid config.
- [ ] Unit: each required var missing in turn → error naming that variable.
- [ ] Unit: two missing vars → one error mentioning both.
- [ ] Unit: malformed CIDR → error.
- [ ] Unit: `SESSION_COOKIE_SECURE` unset → `true`.
- [ ] Unit: the log representation does not contain the password value.

## Done when

- [ ] Starting without `DB_PATH` prints an actionable message and exits non-zero.
- [ ] `.env.example` is complete enough to start the app by copying it.
- [ ] Checkbox ticked in `README.md`.
