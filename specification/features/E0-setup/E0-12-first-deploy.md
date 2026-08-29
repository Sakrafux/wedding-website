# `E0-12` — First deploy to the real server

**Epic:** E0 — Project setup · **Layer:** ops · **Depends on:** `E0-10`

## Story

As an admin, I want the container running on my own server behind my own reverse proxy before any feature exists, so that a deployment problem costs an evening now instead of a crisis in the week before send-out.

This is the exit criterion for M0. It is deliberately the last setup story and deliberately not skippable.

## Scope

**In:**

- The container running on the real server, reachable over HTTPS through the existing reverse proxy.
- `TRUSTED_PROXY_CIDRS` set to the proxy's real network and verified.
- A documented deploy procedure in `README.md`.

**Out:**

- Backup automation and the restore rehearsal → `E-OPS-05`.

## Instructions

1. Record which reverse proxy is in use and its config snippet in the repo. This is an open TODO item, and doing the deploy is what closes it. → Caddy, one site block for the app's own subdomain; snippet in `README.md`.
2. Proxy must forward `X-Forwarded-For` and `X-Forwarded-Proto`, and pass through WebSocket-free plain HTTP to the container port.
3. Set `TRUSTED_PROXY_CIDRS` to the proxy's network, then **verify the app resolves the real client IP** — log a request from a phone on mobile data and confirm the logged IP is the phone's, not the proxy's. Rate limiting is worthless if this is wrong, and it fails silently, so it must be checked by observation rather than assumed.
4. `SESSION_COOKIE_SECURE=true` in production. Confirm the cookie carries `Secure` in the browser once sessions exist.
5. Generate a long random `ADMIN_PASSWORD` and store it in the server's `.env` at mode `0600`.
6. Confirm the security headers from `E0-07` survive the proxy — some proxies strip or duplicate them.
7. Write the deploy procedure down: pull, build, `compose up -d`, check health. Three commands you will otherwise re-derive in six months.

## Test plan

- [x] `https://<subdomain>/api/health` returns 200 from outside the network.
- [x] A deep link returns the SPA, not a proxy 404.
- [ ] Logged client IP matches a real external client, not the proxy. **Deferred to `F1-B05`**: until the trusted-proxy resolution exists, the logged `remoteIP` is the direct peer, which is the proxy by definition.
- [x] Security headers present on the public response.
- [x] Restarting the container preserves the database.
- [x] The site is not reachable over plain HTTP, or is redirected.

## Done when

- [x] The placeholder page is live on the real domain over HTTPS.
- [x] The deploy procedure is written down and was followed, not improvised.
- [x] Checkbox ticked in `README.md`.
