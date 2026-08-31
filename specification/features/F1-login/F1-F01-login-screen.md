# `F1-F01` — Login screen and `CodeInput`

**Epic:** F1 — Household login · **Layer:** frontend · **Depends on:** `F1-B04`

## Story

As a guest holding a printed card, I want to type the code on my phone and get in, so that I can reach the invitation without help.

This is the screen with the highest failure cost in the whole product. If it does not work for a 75-year-old on a five-year-old Android, nothing behind it matters.

## Scope

**In:**

- The unauthenticated landing route at `/`.
- `CodeInput` per [05-design](../../05-design.md).
- Submit, loading, and error states.
- A visible "it does not work" fallback.

**Out:**

- The confirmation screen → `F1-F02`.
- Routing and guards → `F1-F03`.

## Instructions

1. Consume the `F1-B04` contract exactly. Invent no fields.
2. `CodeInput`: one large field, generous height, `font-size` at least 24px. Not six separate boxes — segmented inputs break paste, confuse screen readers, and are a nightmare with an on-screen keyboard.
3. Attributes: `autocapitalize="characters"`, `autocorrect="off"`, `autocomplete="off"`, `spellcheck="false"`, `inputmode="text"`. Autocorrect mangling a 6-character code is a real and infuriating failure.
4. Normalise **as the user types**: uppercase, strip whitespace and dashes, cap at six characters *after* stripping. The card prints the code ungrouped, so a dash is no longer something a guest is shown; stripping one they type anyway costs a character class and spares a rejected code that was typed correctly.
5. Show the hint **outside** the field, as help text — "Sechs Zeichen, genau wie auf der Karte." No placeholder inside the field: a placeholder is not a label and disappears exactly when it would be useful.
6. Real `<label>`: "Dein Code von der Einladungskarte".
7. Submit button is large, full width on mobile, and says something concrete — "Anmelden" — never just an arrow icon.
8. Loading state disables the button and shows a spinner with text. A silent second press on a slow connection is how duplicate submissions happen.
9. Errors appear under the field with an icon, in `danger`, using the German message the API returned verbatim. Do not re-map error text in the frontend — the API owns the wording so it can stay consistent and be fixed without a redeploy of both halves.
10. Under the error, after two failures, reveal the fallback: "Klappt es nicht? Ruf uns an: <Nummer>." Two failures is the point where a person starts to feel stupid, and that is exactly when the way out should appear.
11. On success, invalidate the `me` query and route onward to `F1-F02`.
12. No countdown, no hero, no navigation on this screen. One field, one button.

## Test plan

- [ ] Manual on a real phone: type a code with the on-screen keyboard, in lowercase, and log in.
- [ ] Manual: paste `abc-234` from a message → shown and submitted as `ABC234`, accepted, last digit intact.
- [ ] Component: typing `abc 234` produces the submitted value `ABC234`.
- [ ] Component: an API error renders under the field, with the German text.
- [ ] Component: after two failures, the phone-number fallback is visible.
- [ ] Component: the submit button is disabled while the request is in flight.
- [ ] Accessibility: label is associated with the input; the error is announced (`aria-describedby`, `aria-invalid`).
- [ ] Accessibility: usable at 200% zoom with no horizontal scroll; focus ring visible.

## Done when

- [ ] Someone over 60, who has not seen the app, logs in unaided on their own phone.
- [ ] Checkbox ticked in `README.md`.
