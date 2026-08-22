# `F1-F02` — Household confirmation screen

**Epic:** F1 — Household login · **Layer:** frontend · **Depends on:** `F1-F01`

## Story

As a guest who has just entered a code, I want to be told which household I am now logged in as, so that a mistyped-but-valid code is caught immediately instead of at the point where I have answered on someone else's behalf.

This screen exists for exactly one failure: two valid codes differing by one character. The alphabet makes that unlikely; this makes it harmless.

## Scope

**In:**

- A confirmation screen shown immediately after a successful login.
- "Yes, that's us" continues; "No" logs out and returns to the code entry.

**Out:**

- Any content behind it → F2, F3.

## Instructions

1. Headline names the household: "Willkommen, Familie Müller — seid ihr das?" Use `household.display_name` from the login response.
2. List the known members by first name underneath. Two households with adjacent codes will almost never have the same member names, so the list is what actually catches the error; the household name alone can be ambiguous ("Familie Müller" twice is entirely plausible).
3. Two clear actions, equal weight, both large: **"Ja, das sind wir"** → continue. **"Nein"** → call logout, return to `/`, and show a calm message: "Kein Problem — bitte prüf den Code noch einmal."
4. The "No" path must actually log out server-side, not merely navigate. A session left behind is the bug this screen is meant to prevent.
5. Shown **once per login**, not on every visit. Persist the acknowledgement client-side; a year-long session should not ask this daily.
6. No third option, no "remind me later". Two buttons.

## Test plan

- [ ] Component: renders the household name and member first names from the API response.
- [ ] Component: "Ja" routes onward and does not show again on the next visit.
- [ ] Integration/manual: "Nein" clears the session — `/api/me` afterwards is 401.
- [ ] Component: a household with one member renders sensibly (no empty list, no "und" dangling).
- [ ] Accessibility: both buttons reachable by keyboard, ≥48px tall, and distinguishable without colour.

## Done when

- [ ] Logging in with the wrong-but-valid code is recoverable in one tap.
- [ ] Checkbox ticked in `README.md`.
