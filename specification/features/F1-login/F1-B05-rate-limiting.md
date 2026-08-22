# `F1-B05` — Rate limiting and client IP resolution

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `F1-B04`

## Story

As an admin, I want login attempts limited per IP without ever locking anyone out, so that blind guessing is pointless while a confused guest with fat fingers is never turned away permanently.

## Scope

**In:**

- Trusted-proxy client IP resolution.
- Per-IP limiting on the guest and admin login endpoints.
- A global failure counter for visibility.

**Out:**

- Writing failures to the audit log → `F1-B06`.

## Instructions

1. Client IP: use `X-Forwarded-For`'s rightmost untrusted entry **only** if the direct peer is inside `TRUSTED_PROXY_CIDRS`. Otherwise use the peer address. An unconditionally trusted header makes the limiter bypassable by anyone who sets it — including anyone who has read this repository.
2. If `TRUSTED_PROXY_CIDRS` is empty, never trust the header. Log a warning once at startup, since this is the misconfiguration that silently disables the whole mechanism.
3. Limit: 10 **failures** per hour per IP on the guest endpoint. Successes do not consume budget — a household on a shared connection must not be punished for logging in.
4. Stricter on admin login: 5 per hour. It is the only door where guessing pays.
5. In-memory sliding window or token bucket, keyed by IP. No Redis, no table. State lost on restart is acceptable and even desirable here.
6. Evict idle keys so the map cannot grow without bound.
7. **Never lock out.** The limit produces a 429 with a "try again in a few minutes" message, and it always expires on its own. Per the threat model, locking out a 75-year-old is a worse outcome than the attack being prevented.
8. Set `Retry-After`.
9. Rate-limit responses are 429 with the standard envelope, so the frontend renders them like any other error.

## Contract

Error: `rate_limited` → 429 → "Zu viele Versuche. Bitte warte ein paar Minuten und probier es dann noch einmal. Wenn es weiter nicht klappt, ruf uns einfach an."

## Test plan

- [ ] Unit: peer inside a trusted CIDR + `X-Forwarded-For` → header value used.
- [ ] Unit: peer outside any trusted CIDR + `X-Forwarded-For` → peer used, header ignored.
- [ ] Unit: empty `TRUSTED_PROXY_CIDRS` → header always ignored.
- [ ] Unit: `X-Forwarded-For` with multiple hops → correct entry chosen.
- [ ] Integration: 10 failures → 11th returns 429 with `Retry-After`.
- [ ] Integration: successful logins are not counted.
- [ ] Integration: a different IP is unaffected by the first IP's failures.
- [ ] Integration: admin endpoint limits at 5.
- [ ] Integration: the window expires — after it, requests succeed again. Proves it is a limit and not a lockout.

## Done when

- [ ] Blind guessing is bounded, and no code path can permanently block a household.
- [ ] Checkbox ticked in `README.md`.
