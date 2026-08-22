# `F1-F04` — Admin login screen and shell

**Epic:** F1 — Household login · **Layer:** frontend · **Depends on:** `F1-B07`, `F1-F03`

## Story

As an admin, I want a plain login at `/admin` and a dense shell to hang the admin screens on, so that later admin features are a page each rather than a page plus scaffolding.

## Scope

**In:**

- `/admin/login` with user and password fields.
- Admin layout: sidebar or top nav, placeholders for Haushalte, Dashboard, Sitzplan, Budget, Fotos.
- Logout, and a visible session-expiry affordance.

**Out:**

- Every admin feature page → F5, F6, F7, F8, F9.

## Instructions

1. Not linked from anywhere in the guest UI. Reachable only by typing the URL. Not a security control — `F1-B07` is — but it keeps a curious guest from finding a login form and wondering what is behind it.
2. Standard `<form>` with `autocomplete="username"` / `"current-password"`, so a password manager works. This is the one login in the product that should be in a password manager.
3. Admin layout drops the guest decoration: no serif display sizes, no hero, tighter spacing. Same tokens, different density, per [05-design](../../05-design.md).
4. Admin sessions last 8 hours and do not roll. Show the logged-in state and handle expiry gracefully: a 401 mid-edit returns to the admin login with "Sitzung abgelaufen" — and **must not** drop into the guest login screen, which would be baffling.
5. Nav items for pages that do not exist yet render as disabled placeholders. Better than a 404, and it makes the remaining work visible.
6. Logout clears the session and returns to `/admin/login`.

## Test plan

- [ ] Component: successful login lands on the admin shell.
- [ ] Component: wrong credentials show the German error.
- [ ] Component: 401 mid-session → admin login with the expiry message, not the guest login.
- [ ] Component: a household session visiting `/admin` is refused.
- [ ] Manual: password manager fills the form.
- [ ] Accessibility: labels, focus order, visible focus ring.

## Done when

- [ ] An admin can log in, see the shell, and log out.
- [ ] Checkbox ticked in `README.md`.
