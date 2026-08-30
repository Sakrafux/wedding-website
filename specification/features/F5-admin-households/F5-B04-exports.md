# `F5-B04` — `codes.csv` and `guests.csv` exports

**Epic:** F5 — Admin: households & guests · **Layer:** backend · **Depends on:** `F5-B01`, `F5-B02`

## Story

As an admin, I want the guest list as CSV, so that the print shop can do variable-data printing and so that the whole list is usable in a spreadsheet if the dashboard never ships.

`guests.csv` is the release valve named in [07-roadmap](../../07-roadmap.md): if send-out gets tight, `F6` is dropped and this file does its job.

## Scope

**In:**

- `GET /api/admin/export/codes.csv` — household name and code, for the printer.
- `GET /api/admin/export/guests.csv` — one row per living guest, everything we know.
- A shared CSV writer with the encoding decisions in one place.
- An audit row per export.

**Out:**

- The caterer-shaped export → `F6-B05`. This one is for us, that one is for a vendor and omits what a vendor must not receive.
- Any UI → `F5-F03`.

## Instructions

1. One CSV writer, used by both endpoints, so the encoding decisions below are made once and cannot diverge between the file the printer gets and the file we read.
2. **Encoding: UTF-8 with a BOM, semicolon-delimited, CRLF.** This is a decision about Excel, not about correctness. German Excel splits on the list separator, which is `;` in a German locale, so a comma-delimited file lands entirely in column A; and without a BOM it reads UTF-8 as Latin-1, so `Müller` becomes `MÃ¼ller`. Both of these are silent, both would be discovered by the print shop rather than by us, and one of them ends up on eighty cards. Record the trade-off: the file is then not RFC 4180 and not what `pandas.read_csv` expects by default.
3. Quote every field. Cheap, and it removes the class of bug where a name containing a semicolon splits a row.
4. `Content-Type: text/csv; charset=utf-8` and `Content-Disposition: attachment; filename="codes.csv"`. A CSV that renders in the browser instead of downloading is a CSV somebody copies out of the page by hand.
5. `codes.csv` columns: `haushalt;code`. The code goes in the **printed** form (`ABC-234`) — that is the string that must appear on the card, and asking the print shop to insert a dash is asking for the one they forget. German column headers here, uniquely, because a print shop reads them.
6. `guests.csv` is a **database output**: every column of `guest`, plus every column of the household that owns it, one row per guest. Not a curated subset — this file is the release valve, and the point of a release valve is that it does not require anyone to have guessed correctly in advance which field would be wanted.
   Headers are the column names verbatim, English, prefixed `household_` where they come from the household — `guest_id`, `household_id`, `household_display_name`, `household_code`, `first_name`, … The prefix is what stops `created_at` from being ambiguous, and matching the schema exactly means a question about a value is answered by reading `03-data-model` rather than by guessing what a friendly header meant.
7. Include **soft-deleted guests**, with `deleted_at` as the third column so it cannot be missed while scanning. Excluding them would make the file disagree with the database it claims to dump, and a removed plus-one is exactly the row somebody eventually wants to see.
   The cost is real and must be stated in the file's own description (`F5-F03`) as well as here: **this file is a dump, not a headcount.** Anything counted from it has to filter `deleted_at` first. `F6-B05`'s caterer export is the one that has already done that.
8. The RSVP columns are empty until `F3` fills them. Emit them anyway: the file's shape should not change on the day the answers start arriving, or every spreadsheet built on it breaks at once.
9. Stream the rows to the response writer rather than building the file in memory. Not for the sixty rows here — for the habit, and because the alternative silently sets a precedent for the photo ZIP in `F10-B03`.
10. Record each export as an **info log line** with the row count, not as an `audit_log` row. `codes.csv` is the whole key list leaving the server and deserves a trace, but `audit_log.action` has no `read` value, and adding one is a migration that would also invite every future read to be logged there — the `CHECK` constraint is what keeps that table a record of *changes* rather than a general event log. Reconsider if a read ever needs to be correlated with a change; today nothing does.
11. `codes.csv` is the most sensitive artefact this application produces. `E-OPS-07` deletes it from both ends after send-out; nothing in the app can enforce that, so the frontend says so at the point of download (`F5-F03`).

## Contract

```http
GET /api/admin/export/codes.csv
```

Response `200`, `text/csv`:

```csv
"haushalt";"code"
"Familie Müller";"ABC-234"
```

```http
GET /api/admin/export/guests.csv
```

Response `200`, `text/csv`, one row per living guest.

Errors: `unauthenticated` → 401

## Test plan

- [ ] Integration: `codes.csv` has a header row and one row per household, with codes in printed form.
- [ ] Integration: the body starts with a UTF-8 BOM and uses `;` and CRLF.
- [ ] Integration: a household name containing a semicolon and a quote round-trips through a CSV reader unmangled.
- [ ] Integration: umlauts survive — assert the bytes, not just the string, since this is exactly what the BOM is for.
- [ ] Integration: `Content-Disposition` is `attachment` with the expected filename.
- [ ] Integration: `guests.csv` carries every column of `guest` and of `household` — assert against the column list, so a migration that adds a field and forgets this file fails here.
- [ ] Integration: soft-deleted guests are present, with `deleted_at` populated and third in the row.
- [ ] Integration: the RSVP columns are present and empty before `F3`.
- [ ] Integration: household-session and anonymous requests to both routes → 401. These two files are the largest disclosure in the product.
- [ ] Integration: an export writes a log line naming the row count.

## Done when

- [ ] `codes.csv` opens correctly in German Excel, with umlauts intact and one code per column.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
