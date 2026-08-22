# `E0-07` — Security headers, CSP, noindex

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-01`

## Story

As an admin, I want the site to send strict headers and to stay out of search engines, so that "nothing is publicly readable, nothing is indexed" is enforced by the server rather than by hope.

## Scope

**In:**

- Middleware setting the header set from [06-privacy-security](../../06-privacy-security.md) on every response.
- `robots.txt` with a blanket disallow.

**Out:**

- TLS and HSTS — the reverse proxy's job.

## Instructions

1. Set on every response:
   - `Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self'; font-src 'self'; script-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'`
   - `X-Content-Type-Options: nosniff`
   - `Referrer-Policy: no-referrer`
   - `X-Frame-Options: DENY`
   - `Permissions-Policy: geolocation=(), microphone=(), camera=()`
   - `X-Robots-Tag: noindex, nofollow`
2. Serve `robots.txt` with `User-agent: *` / `Disallow: /`.
3. Verify the CSP against the built frontend **during `E0-08`/`E0-09`**, not later. Vite's dev server and Tailwind are fine with this policy, but a stray inline style or an unexpected data URI will only appear once the real bundle is served — and a CSP that gets loosened in a panic on deploy day never gets tightened again.
4. No `unsafe-inline`, no `unsafe-eval`. If something needs them, the something is wrong.

## Test plan

- [ ] Integration: all six headers present on an API response.
- [ ] Integration: all six headers present on the SPA `index.html` response.
- [ ] Integration: `robots.txt` returns the disallow.
- [ ] Manual: the built frontend loads with zero CSP violations in the browser console.

## Done when

- [ ] Headers verified on the deployed container, through the real reverse proxy — proxies sometimes strip or duplicate headers.
- [ ] Checkbox ticked in `README.md`.
