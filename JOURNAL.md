# Journal

Work log for the wedding web app. Newest entry first. One `##` heading per day, then: what the day was about, a `Time:` line, a `Cost:` line per model used.

## 2026-08-25

First code. `E0-01` done: Go module `github.com/Sakrafux/wedding-website` on Go 1.26, the full package tree from `04-architecture.md` with a `doc.go` in each package carrying that package's rules, chi router with `httplog`, `GET /api/health` returning `{"status":"ok"}`, and graceful shutdown on SIGINT/SIGTERM with all four server timeouts set explicitly. Direct dependencies held to `chi/v5` and `httplog/v2` as the story asks; the integration test uses stdlib `httptest` only, since testify arrives with the real harness in `E0-11`.

The web adapter got split further than `04-architecture.md` originally drew it: `web` now holds only the router, with `handler/`, `httpio/`, `dto/` and `middleware/` beneath it. The forcing constraint is the error envelope — `middleware` has to reject requests in the same shape handlers use, and since `web` imports `middleware` to build the router, envelope writers inside `web` would be an import cycle. `httpio` owns them, so a 401 from the session gate and a 404 from a handler are identical by construction. Its name cost a detour: `helper/` was rejected because it excludes nothing and therefore collects everything, and a shorter `respond.JSON` / `respond.Error` pair was tried and dropped again — a Go function called `Error` reads as the `error` interface method, and `Fail` reads as `testing.T`. The explicit `WriteJSON` / `WriteError` cannot be misread. Spec tree and the deviations section updated to match.

Two more decisions worth recording. `middleware.RealIP` is deliberately not installed — it believes `X-Forwarded-For` from anyone, which would make the login rate limit bypassable, so `F1-B05` will resolve the client IP against `TRUSTED_PROXY_CIDRS` instead. And the error envelope from `04-architecture.md` (`{"error":{"code","message"}}`) already exists in `web/dto` in minimal form, because `/api` needs a JSON 404 today; `E0-06` grows it into the shared domain-error mapping rather than inventing a second shape.

Added one convention to `CLAUDE.md` off the back of this: comments that point at future work must name the story (`E0-06 replaces this`), and a story is not ticked until `grep -rn "<ID>" --include='*.go' .` comes back empty. This code already carries seven such pointers, and they are the comments most likely to rot — the code around them stays correct, so nothing draws the eye to them.

Also confirmed the repo layout against a `backend/` + `frontend/` split and kept the spec's root layout: `go:embed` cannot cross the module root, so splitting would force a copy of `frontend/dist` into the Go tree at build time and allow a stale copy in development.

Time: 1h

Cost: Opus 5 — $4.11

## 2026-08-22

Finished the rough plan: wrote `05-design.md` (design system, colours, type, German enum labels), `06-privacy-security.md` (threat model, guest-vs-admin data boundaries) and `07-roadmap.md` (undated milestones, invitations Oct/Nov 2026 as the real deadline). Started the per-feature spec split under `specification/features/` with an index README and `_TEMPLATE.md`. Settled on this journal as the place to track effort and AI spend.

Several facts firmed up along the way: 2027-07-17 as the working wedding date (venue availability decides it, and the venue is being booked within two weeks), the print shop confirmed for variable-data printing, and a separate save-the-date rejected — at nine months out with a mostly local guest list, one mailing does both jobs, and the reasoning is recorded in `02-features.md`. Toned down the privacy document afterwards: the database-leak and photo-EXIF sections were written for a threat model stricter than the one we actually have, so both now read as proportionate hygiene rather than key management, with the photo gallery framed as a shared album carrying about as much responsibility as a family group chat.

Then populated the backlog properly: 23 story files written in full — `E0-setup` (12 stories, ending at a deploy to the real server before any feature exists) and `F1-login` (11 stories, backend-then-frontend per slice). Every other epic through F10 sits in the index as bare checkboxes, plus an `E-ops` epic so the non-code gates — print run, restore rehearsal, send-out, deadline flip, wind-down — are tracked like everything else. Switched the frontend package manager to pnpm, moved `TODO.md` to the repo root and split its scope from the build tracker, and wrote the root `README.md`.

Time: 4.5h

Cost: Opus 5 (1M context) — $11.31

## 2026-08-21

Set the project up from scratch and locked in the big decisions: Go single binary with embedded React/Vite frontend, SQLite, trimmed hexagonal layout. Wrote `CLAUDE.md` plus the first spec batch — `01-vision-scope.md`, `02-features.md`, `03-data-model.md`, `04-architecture.md` — and the initial TODO list.

Time: 3h

Cost: Opus 5 (1M context) — $6.57
