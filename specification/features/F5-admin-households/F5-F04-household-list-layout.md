# `F5-F04` — Household list: actions above, table last

**Epic:** F5 — Admin: households & guests · **Layer:** frontend · **Depends on:** `F5-F01`

## Story

As an admin, I want the controls to stay where I left them as the guest list grows, so that creating the sixtieth household does not mean scrolling past fifty-nine rows.

## Scope

**In:**

- Page order: heading, counts, the one-row create form, search and filters, the table, the exports.
- A second filter: households that have not answered.
- A sticky table header.

**Out:**

- Pagination. Rejected on 2026-09-01: sixty rows with a search box and two filters is a screen you scan, and a pager would hide the rows the filters just found.
- Any RSVP counts beyond the existing column → `F6`.

## Instructions

1. The table goes last because it is the only element that grows. Everything above it is fixed in height, so every control keeps its position for the life of the project.
2. The exports stay at the bottom, below the table: two links used twice in the project's life, and the end of the table is where you already are when you go looking for them. The code warning stays glued to the codes link — it is the most sensitive artefact the app produces, and a bare link in a toolbar is how it stops being read.
3. Create stays an inline one-row form rather than a dialog: it is the control used sixty times in a single seeding sitting (`E-OPS-01`), and it already navigates to the new household's detail page, so nothing needs finding afterwards.
4. Filters sit directly above the table they filter, not in the action strip. "What I do" and "what I see" are different rows.
5. The new filter reads `rsvp_submitted_at`, and combines with the never-logged-in filter as an AND — two filters that quietly ORed would produce a list nobody can explain.
6. Sticky header via `position: sticky` on the `<thead>` cells, with the surface colour set: a transparent sticky header shows rows sliding under the words.

## Test plan

- [ ] Component: the create form renders above the table and the exports below it.
- [ ] Component: the "ohne Antwort" filter shows exactly the households with no `rsvp_submitted_at`.
- [ ] Component: both filters together show only households matching both.
- [ ] Component: search still narrows the rows, and the empty-result sentence still appears.

## Done when

- [ ] The page reads action-first, and no control moves as households are added.
- [ ] Checkbox ticked in `README.md`.
