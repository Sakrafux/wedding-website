# `F11-04` — A navigation between pages is not a page load

**Epic:** F11 — Cross-cutting quality · **Layer:** frontend · **Depends on:** `F1-F03`, `F2-F01`

## Story

As a guest, I want moving between pages to show me the next page, so that tapping "Ablauf" does not flash a grey skeleton of a login screen at me.

## Scope

**In:**

- The `/_guest` guard resolves synchronously when the session is already cached.
- A content-shaped pending state that renders **inside** the navigation, for the routes that genuinely load data.
- `RoutePending` reserved for the cold load, where its shape is right.

**Out:**

- Replacing skeletons with a spinner everywhere. Considered on 2026-09-01 and rejected: the flash was not the skeleton's fault, it was a guard that awaited a value it already had, and a spinner on every navigation would be the same interruption with less information.

## Instructions

1. The cause: `beforeLoad` on `/_guest` is `async` and awaits `ensureQueryData(me)`. An `async` guard puts the router into its pending state for at least one tick even when the answer is in the cache, and `defaultPendingMs: 0` renders that state immediately. Every navigation between hardcoded content pages therefore showed a full-screen skeleton. Nothing was being fetched and nothing was being code-split.
2. Read the cache first with `getQueryData` and return **without a promise** when it holds a value — `undefined` means "not fetched", and a cached `null` means "not logged in", which is a value. Only the cold load awaits.
3. `defaultPendingMs: 0` stays. It is right for the case it was written for: on a cold load the guard genuinely has to wait, and an unstyled gap before the skeleton reads as a broken page.
4. `/zusagen` keeps its loader — it has real data to fetch — and gets a pending component shaped like the page, rendered inside the guest chrome. `RoutePending` centres a login-shaped column in the viewport, which is correct at `/` and wrong under a navigation bar.
5. Same cache-first treatment for that loader, so returning to `/zusagen` with the answer already loaded renders it immediately.

## Test plan

- [ ] Component: navigating between two content pages with `me` cached renders no pending state.
- [ ] Component: a cold load with no cached session shows `RoutePending`.
- [ ] Component: `/zusagen` with an empty cache shows the in-chrome pending state, not the full-screen one.
- [ ] Component: an unauthenticated cold load still redirects to `/` with the intended path.

## Done when

- [ ] Moving through the site looks like moving through a site.
- [ ] Checkbox ticked in `README.md`.
