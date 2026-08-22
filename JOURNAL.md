# Journal

Work log for the wedding web app. Newest entry first. One `##` heading per day, then: what the day was about, a `Time:` line, a `Cost:` line per model used.

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
