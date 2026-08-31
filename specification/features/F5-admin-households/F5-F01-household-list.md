# `F5-F01` — Household list

**Epic:** F5 — Admin: households & guests · **Layer:** frontend · **Depends on:** `F5-B01`, `F1-F04`

## Story

As an admin, I want one screen listing every household with its code and whether they have logged in, so that I can see the state of the guest list at a glance instead of reconstructing it from a spreadsheet.

## Scope

**In:**

- `/admin/haushalte`, replacing the disabled placeholder in the admin nav.
- A dense table: name, code, member count, last login, RSVP state.
- Search by name, and a filter for "hat sich nie angemeldet".
- "Haushalt anlegen".

**Out:**

- Editing a household or its members → `F5-F02`.
- Export buttons → `F5-F03`.
- Any RSVP counts or charts → `F6`.

## Instructions

1. Consume the `F5-B01` contract exactly. Invent no fields.
2. A real `<table>`, not a grid of `<div>`s. This is tabular data, the admin will sort and scan it, and a screen reader announcing row and column headers is worth more here than any layout convenience.
3. Columns: Haushalt, Code, Personen, Letzte Anmeldung, RSVP. Codes render exactly as printed (`ABC234`) in a monospaced-feel style, because their only use on this screen is being read aloud or compared against a card.
4. `last_login_at` empty renders as a visible marker, not a blank cell — a blank reads as "not loaded". This column is the answer to "did they even see it?", which is what drives the nudge calls before send-out.
5. Colour is never the only signal (`05-design`): "nie angemeldet" carries a text label or an icon as well as its colour.
6. Dates in German short form, and a relative hint for recent ones ("vor 3 Tagen"). The absolute date is what gets read out on the phone; the relative one is what makes the column scannable.
7. Search filters client-side. Sixty rows are all in memory already, and a server round trip per keystroke is latency for nothing.
8. Sort by name by default, matching the server's order, so a reload does not reshuffle the screen.
9. Numbers use tabular figures — the token is already set globally on `table`.
10. "Haushalt anlegen" opens a small form (name only) and navigates to the new household's detail page, where the code is now visible and the members can be added. Creating and then hunting for the row you just created is the flow to avoid.
11. The count of households and the count of never-logged-in households appear above the table as plain text. Not stat tiles — those are `F6`'s, and two numbers do not need a component.

## Test plan

- [ ] Component: rows render from the API response, with codes in printed form.
- [ ] Component: a household that has never logged in shows the marker; one that has shows the date.
- [ ] Component: search narrows the rows and is case-insensitive.
- [ ] Component: the "nie angemeldet" filter shows exactly those rows.
- [ ] Component: creating a household navigates to its detail page.
- [ ] Component: a 401 from the list query lands on the **admin** login, not the guest one.
- [ ] Accessibility: the table has a caption or an accessible name, and column headers are real `<th scope="col">`.

## Done when

- [ ] The full guest list is legible on one screen, and the households to chase are obvious without counting.
- [ ] Checkbox ticked in `README.md`.
