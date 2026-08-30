# `F5-F03` — Export actions

**Epic:** F5 — Admin: households & guests · **Layer:** frontend · **Depends on:** `F5-B04`, `F5-F01`

## Story

As an admin, I want to download the code list and the guest list from the admin UI, so that sending the print shop their file does not involve a terminal.

## Scope

**In:**

- Download buttons for `codes.csv` and `guests.csv`.
- A warning at the point of download about what `codes.csv` is.

**Out:**

- The caterer export → `F6-F03`.
- Any transformation of the file in the browser. The server produces the bytes; the browser saves them.

## Instructions

1. Plain `<a download>` links to the endpoints, not a fetch-and-blob dance. The cookie rides along, the browser handles the save dialog, and there is no in-memory copy of the code list in a tab that stays open all afternoon.
2. Label each with what it is for, not with its filename: "Codes für die Druckerei" and "Gästeliste (Rohdaten)". Nobody downloading these thinks in filenames.
3. Say next to the guest list that it includes entfernte Personen and is not a headcount. It is a dump of the table, soft-deleted rows included; somebody will otherwise count its rows in May and brief the caterer from the total.
4. Next to `codes.csv`, one sentence: this file contains every login code, and it should be deleted from both ends once the cards are printed (`E-OPS-07`). The app cannot enforce that, and a warning nobody sees is the same as no warning.
5. Show the household count next to the codes link, so a truncated or empty file is obvious before it reaches the printer rather than after.
6. Both actions live on the household list page (`F5-F01`). A separate exports page for two links is a page nobody would find.
7. No progress UI. Sixty rows are instantaneous; a spinner would be a lie about the work involved.

## Test plan

- [ ] Component: both links point at the right endpoints and carry the `download` attribute.
- [ ] Component: the `codes.csv` warning is present and mentions deleting the file.
- [ ] Manual: downloading `codes.csv` and opening it in Excel shows two columns, umlauts intact, one row per household.
- [ ] Accessibility: the links are reachable by keyboard and their accessible names say what the file is.

## Done when

- [ ] The print shop's file can be produced and checked without a terminal.
- [ ] Checkbox ticked in `README.md`.
