# `F1-F03` — App shell, routing, query client

**Epic:** F1 — Household login · **Layer:** frontend · **Depends on:** `F1-B04`, `F1-F02`

## Story

As a guest, I want the site to remember I am logged in and to take me straight to the content, so that opening the link months later just works.

## Scope

**In:**

- TanStack Router route tree: public login route, authenticated layout, admin layout.
- TanStack Query client with the `me` query as the session source of truth.
- Route guards driven by `me`.
- Global 401 handling.
- Loading and error boundaries.

**Out:**

- Navigation chrome and the bottom bar → `F2-F01`.
- Admin screens → `F1-F04`.

## Instructions

1. `me` is a single query, fetched once at app start, and is the **only** place session state lives. No auth state duplicated in a store or context — two sources of truth for "am I logged in" is a reliable source of bugs.
2. Guarded routes redirect to `/` when `me` is 401, preserving the intended path so the guest lands where they were headed after logging in.
3. A global response handler: any 401 from any query invalidates `me` and drops the user to the login screen with a plain message ("Du wurdest abgemeldet."). A year-long session can still be revoked, and the app must handle that without a white screen.
4. `staleTime` on `me` of a few minutes; the flags it carries (`rsvp_open`, `seating_published`) change rarely, and refetching on every focus is pointless traffic.
5. While `me` is loading, render a skeleton — never a blank page and never a flash of the login screen for an already-authenticated guest. That flash reads as "it logged me out again", which erodes trust in exactly the audience that has least of it.
6. Error boundary rendering the German error with the request ID from the envelope, plus a "nochmal versuchen" button.
7. Admin routes live under `/admin` behind their own guard checking `subject_type === 'admin'`.
8. Handle offline and network failure distinctly from 401: "Keine Verbindung" with a retry, not a logout. Dropping a guest to the login screen because a train went through a tunnel is a bad trade.

## Test plan

- [ ] Component: unauthenticated visit to a guarded route → login screen.
- [ ] Component: after login, redirect goes to the originally requested path.
- [ ] Component: a 401 mid-session → login screen with the logged-out message.
- [ ] Component: a network error → retry UI, **not** a logout.
- [ ] Component: no flash of the login screen while `me` is loading.
- [ ] Component: household session cannot reach an `/admin` route.
- [ ] Manual: hard refresh on a deep link keeps the session and the route.

## Done when

- [ ] Closing the browser and returning weeks later lands straight on the content.
- [ ] Checkbox ticked in `README.md`.
