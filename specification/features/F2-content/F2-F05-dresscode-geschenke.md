# `F2-F05` — Dresscode, Geschenke

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F01`

## Story

As a guest, I want to know what to wear and what you would like as a present, so that I do not have to ask somebody else in the family.

## Scope

**In:**

- `/dresscode` and `/geschenke`, two pages, one story.
- `InfoSection`: the generic heading + prose block both pages are built from, and every later content page reuses.
- The gift wording and the bank details, which are published (see instruction 6).

**Out:**

- A gift registry with purchase tracking — out of scope in [01-vision-scope](01-vision-scope.md), and it stays out.
- FAQ and contact → `F2-F06`.

## Instructions

1. Two routes, one story: both pages are a heading and three paragraphs, and building `InfoSection` twice would be the only work in splitting them. Both are reachable from `/mehr`, not from the bottom bar.
2. `InfoSection` is the reusable piece this story leaves behind — heading, optional lead sentence, prose at `max-w-prose`. `F2-F06` and `F2-F07` use it rather than inventing their own.
3. Dress code says what to wear in plain words and gives an example, not a category name. "Festlich" means five different things to five relatives; a sentence describing what we will be wearing settles it.
4. Say what the ground and the weather will be like if it matters — an outdoor lawn in July is a shoe decision, and it is the question we would otherwise be asked by phone.
5. Gift wording is ours to write, informal, short, and free of the word "Geldgeschenk" doing all the work alone. Guests who want to bring something should not feel corrected.
6. **Bank details are a publishing decision, not a layout one.** Everything on a content page is compiled into the JavaScript bundle, and the bundle is served to anyone who requests the site — the session gate covers `/api`, not the SPA (`internal/infrastructure/web/static.go`, `E0-09`). An IBAN on this page is therefore readable by anyone who knows the URL, and the site is unindexed (`E0-07`) but not secret. **Decided 2026-08-31: publish it.** Semi-public is acceptable — an IBAN lets somebody send money, not take it, the site is unindexed (`E0-07`), and "IBAN auf Anfrage" costs every guest a phone call to do the thing we asked them to do. What is not acceptable is publishing it *believing* it is behind the login, so the reasoning above stays on record.
7. Render the IBAN in tabular figures, grouped in fours, as selectable text with a copy button. Nobody should transcribe an IBAN from a phone screen by hand.
8. No image on either page. Both answer a question in a sentence; a decorative photo pushes the answer below the fold.
9. **Open input, tracked in [TODO.md](../../../TODO.md):** the dress code wording, the gift wishes, and the account details themselves (IBAN and account holder).

## Test plan

- [ ] Component: `renderApp("/dresscode")` and `renderApp("/geschenke")` each render their heading and body from `labels.ts`.
- [ ] Component: both pages are reachable from `/mehr`.
- [ ] Component: the copy button copies the unspaced IBAN, and the visible grouping is not what lands on the clipboard.
- [ ] Accessibility: one `<h1>` per page, prose at the capped measure, and the copy button has a real accessible name and a live-region confirmation.

## Done when

- [ ] Both questions are answered on the site, and an IBAN can be copied from a phone without transcribing it.
- [ ] Checkbox ticked in `README.md`.
