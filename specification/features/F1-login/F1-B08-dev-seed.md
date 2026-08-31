# `F1-B08` — `cmd/seed`: local development households

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** F1-B01, E0-03, E0-04, E0-05

## Story

As a developer, I want a command that inserts a few households with real login codes into my local database so that I can exercise the login flow without hand-writing SQL against a live WAL-mode SQLite file.

Until this exists the only code that creates a household is `tests/integration/fixtures_test.go`, which is unreachable from any binary — a fresh `make run` therefore serves a login screen that no code can get past.

## Scope

**In:**

- `cmd/seed`, a second binary in the same module, invoked through `make seed`.
- `-households N` (default `1`) and `-guests N` (default `2`) — how many households to insert and how many adult members each gets.
- Codes come from `domain.GenerateCode`. **Not** fixed and not derived from the counter: the tool prints them, and a seeded code that is shaped differently from a printed one would let a bug in normalisation or in the input mask survive local testing.
- Prints one line per household to stdout — id, display name, code in `FormatCode` display form — because a plaintext code that is never printed is a household nobody can log in as.
- Reads `DB_PATH` from the same environment the app does, and runs the migrations before inserting, so it works against an empty file.
- Loud, unambiguous documentation that this is a local development tool: doc comment on the command, a banner line in its output, and the `README.md` dev section.

**Out:**

- Any production or admin path for creating households — that is `F5-B01` (CRUD endpoints) and `F5-B03` (code generation and regeneration per household). This command is scaffolding for the dev shell and is not a substitute for either.
- RSVP answers, guest-added members, children, seating, budget and photo fixtures. Households and adult members only; the later epics own their own data and there is nothing to look at yet.
- Deleting or resetting. `rm local/wedding.db*` is shorter than a flag, and a `--reset` on a tool that reads `DB_PATH` from the environment is one shell mistake away from the deployed volume.
- A hard runtime guard against running it in production. It is documented, not enforced: the binary is not in the image (`Dockerfile` builds `./cmd/wedding`), so there is nothing there to run.

## Instructions

1. `cmd/seed/main.go`. Package doc comment states the development-only purpose first.
2. Flags `-households` and `-guests`, validated: below 1 is a usage error, not a silent no-op.
3. `configuration.Load()` for the environment, `configuration.OpenDatabase` for the handles, `persistence.Migrate` before the first insert.
4. Insert with plain SQL on `database.Write`, one transaction per household, so an interrupted run leaves whole households rather than a household with half its members. The SQL lives in `cmd/seed` on purpose and carries a `F5-B01` forward reference: once `HouseholdStore` grows a create method, this command calls it instead.
5. Display names are obviously synthetic (`Familie Testhaushalt <n>`, members `<Vorname> Testhaushalt <n>`), numbered from the current household count so a second run does not repeat the first one's names. Deriving a member's name from the household's is something only this command may do — it invented the household name in the first place, whereas a real one is free text.
6. Retry a `UNIQUE` collision on `household.code` a few times before failing. Collisions are ~impossible at 32^6, but the retry is what makes the UNIQUE index the sole authority on uniqueness — the tool never queries for a free code first.
7. Exit non-zero with a plain stderr message on any failure. No structured logger: this is a shell tool, and its output is read by a person.

## Contract

None. No HTTP surface, no DTO — the command talks to the database and to stdout.

## Test plan

- [ ] unit/integration (`cmd/seed`, temp-file SQLite): the requested number of households and members exist afterwards, and every code passes `domain.ValidateCode`.
- [ ] a second run adds to the first rather than colliding on names or codes.
- [ ] negative: `-households 0` fails with a usage error and writes nothing.
- [ ] the codes it prints are accepted by `POST /api/auth/login` — the point of the tool, and the one thing a wrong storage form would break.

## Done when

- [ ] `make seed` against a fresh `local/wedding.db` prints usable codes, and one of them logs in at `http://localhost:5173/hochzeit/`.
- [ ] `README.md`'s local development section documents the command and says it is dev-only.
- [ ] Tests above pass; `go test ./...` is green
- [ ] Checkbox ticked in `README.md`
