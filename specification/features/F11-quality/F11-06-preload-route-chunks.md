# `F11-06` — Every route's chunk is already there when it is tapped

**Epic:** F11 — Cross-cutting quality · **Layer:** frontend · **Depends on:** `F11-04`

## Story

As a guest, I want tapping a page to show that page, so that a site of a dozen short pages never makes me wait for one.

## Scope

**In:**

- `defaultPreload: "render"` — the route behind every rendered `<Link>` is fetched ahead of the tap.
- `defaultPendingMs: 150`, so a transition that finishes under that renders nothing at all.
- `SectionPending` as the default pending state; `RoutePending` kept for the three places where the whole app really is booting.
- One router factory, shared by `main.tsx` and the test harness.

**Out:**

- Turning `autoCodeSplitting` off. Considered on 2026-09-01 and rejected: it removes the per-route fetch, but `RSVPForm`'s chunk is ~24 kB gzipped of Radix dialog code and would land in the initial bundle — paid on the login screen by every guest, including the ones who never open the form that visit.
- The guard fix, which was a different cause with the same symptom → `F11-04`.

## Instructions

1. `autoCodeSplitting: true` (`vite.config.ts`) puts every route component in its own chunk, so the first visit to a page fetches one — and with `defaultPendingMs: 0` the skeleton rendered for exactly as long as that took. Real, and separate from `F11-04`'s guard: fixing the guard removed the flash on *repeat* navigations only.
2. `"render"` rather than `"intent"`: the navigation renders a link to every guest page on every page, so `render` warms all of them once the shell is up, while `intent` still races a fast tap. The chunks are 0.2–1.3 kB gzipped each — a handful of kB, fetched after first paint rather than before it.
3. Preloading runs a route's loader as well as its chunk. That is wanted here and costs nothing: `/zusagen`'s loader is cache-first (`F11-04`) and the chrome already reads the RSVP answer for its own nav label.
4. `defaultPendingMs: 150` is the actual flash-killer, and covers what preloading cannot — the admin routes the nav does not link exhaustively, and a cold cache. It does not contradict `F11-04`'s reasoning for `0`: that was about the cold load, where the wait is long enough that 150 ms is invisible.
5. Pending components resolve **per route**, falling back to the router default — a layout route's does not cover its children (`Match.js`). So the default becomes `SectionPending`, and the three routes that are genuinely full-screen say so explicitly: `_guest` (the guard, before any chrome exists), `/` and `/admin/login`.
6. `main.tsx` and `src/test/render.tsx` each built a router with their own copy of these options, which is how a test would keep passing against settings the app no longer has. One `createAppRouter` now, differing only in the history it is given.

## Test plan

- [ ] Component: the shared factory's router carries `defaultPreload: "render"` and `defaultPendingMs: 150` — a guard against these being dropped again, since both are invisible until somebody reports a flash.
- [ ] Component: the existing suites keep passing against the shared factory, including the preload-triggered loader calls.
- [ ] Manual: throttled to Slow 3G, tapping through the guest nav shows no skeleton after the first paint.

## Done when

- [ ] Moving between guest pages fetches nothing at tap time, and no pending state appears for a transition a person would call instant.
- [ ] Checkbox ticked in `README.md`.
