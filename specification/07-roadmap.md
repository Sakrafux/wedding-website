# 07 — Roadmap

Status: draft · Last updated: 2026-08-22

## Anchors

Three fixed points. Everything else is ordering, not scheduling.

| Anchor | When |
|---|---|
| Wedding | **2027-07-17** — working assumption. Venue and church are being fixed within ~2 weeks of 2026-08-22, and the venue's availability sets the date, so this becomes firm in early September 2026. Any change is within days and costs a copy edit |
| Invitation send-out | **October / November 2026** |
| RSVP deadline | **~2 months before the wedding**, so around mid-May 2027 |

Development target: **finished by the end of 2026, ideally September/October** — which is to say, done in time for send-out rather than done at leisure.

Milestones below are deliberately **undated**. The order is the plan; assigning calendar dates to a hobby project this small produces a schedule that is wrong within a fortnight and then gets ignored. What matters is what must be true before send-out and what can follow it.

Print shop confirmed to do variable-data printing, so the F1 design stands as specified — individually printed codes, generic QR.

## The one thing that matters

**Send-out is the real deadline, not the wedding.** After the cards are in the post the guest list is live, the codes are printed, and the RSVP field set is frozen. Before that date, everything is cheap to change; after it, some things cost a reprint.

So the question for every milestone is only: *does this have to work before the cards go out?*

**Must ship before send-out:**

| Feature | Why it cannot wait |
|---|---|
| F1 Household login | The code on the card is worthless without it |
| F5 Admin households & guests | Codes have to be generated and exported before they can be printed |
| F3 RSVP | The reason the card points anywhere |
| F4 Plus-ones & children | Part of the same form; adding it later is a migration against live data |
| F2 Informational content | A guest who logs in on day one and finds an empty site does not come back |
| F11 Accessibility basics | Retrofitting is worse than building it in, and this audience needs it most |

**Can follow send-out, safely:**

| Feature | Why the delay is free |
|---|---|
| F6 Admin RSVP dashboard | On send-out day nobody has answered yet. A `guests.csv` export covers the first weeks; the dashboard can land while the answers trickle in |
| F8 Budget | Admin-only, no guest ever sees it, no dependency on anything |
| F9 Curated gallery | Nice to have, gated by `gallery_visible` |
| F7 Seating | Cannot start until the RSVP deadline anyway |
| F10 Guest uploads | Nothing to upload before the wedding |

**F6 is the release valve.** If the calendar gets tight, ship send-out without the dashboard rather than pushing send-out. That decision is available right up to the last week.

## Milestones

### M0 — Foundations

Nothing user-visible. The point is that every later milestone is writing features rather than fighting infrastructure.

- Go module, `cmd/wedding`, the package layout from [04-architecture](04-architecture.md), chi router, `httplog`.
- Config from env with hard failure on missing required vars.
- SQLite open with the four pragmas, single-writer pool, read pool.
- Embedded migration runner + `schema_migration`, with migration `0001` creating the full schema from [03-data-model](03-data-model.md).
- Vite + TS + Tailwind + shadcn scaffold; `go:embed` of `dist/` and the SPA fallback.
- Dockerfile (multi-stage, distroless, non-root), Compose file, volume paths.
- Integration test harness: temp-file SQLite, migrations applied, one trivial endpoint asserted.

**Exit criterion:** the container runs on the real server, behind the real reverse proxy, and serves a "hello" page over HTTPS.

Deploy on day one, not at the end. A deployment problem found now is an evening; found the week before send-out it is a crisis.

### M1 — Login and the guest list

F1 and F5 together, because a code you cannot generate is not testable and a household list with no login is not usable.

- Code generation (`crypto/rand`, 32-char alphabet, collision retry), normalisation, `CodeInput`.
- Session issue / validate / refresh / revoke; the middleware; rate limiting with the trusted-proxy rule.
- Household confirmation screen.
- Admin login, admin gate, `/api/admin` rejection for household sessions.
- Admin CRUD for households and guests; `codes.csv` export.
- Integration tests: login happy path, normalisation variants, rate limit, admin boundary, and the assertion that no guest response contains `code` or `admin_note`.

**Exit criterion:** a code generated in admin can be typed on a phone and produces a year-long session — verified by someone who did not write the code.

### M2 — RSVP

The heart of the product, and the part that must not change afterwards.

- F3: the RSVP form, scope selector, per-member overrides, scope-gated catering fields, transport steppers, household note.
- F4: guest-added members, soft cap, deletion rules.
- Domain unit tests for every invariant in [04-architecture](04-architecture.md).
- Audit logging on every RSVP mutation.
- Deadline behaviour: read-only rendering after `rsvp_deadline`.

**Exit criterion — Gate 1: the RSVP field set is frozen.** Write the F3 story before building it; the field set grew three times during planning, and every later addition is a migration against live guest data.

Before freezing, one deliberate dry run: fill the form as a household containing an infant, a church-only grandmother, a vegan plus-one, and someone arriving after dinner. If all four are expressible without a free-text workaround, the field set is right.

### M3 — Content and quality

- The design system from [05-design](05-design.md): tokens, fonts, type scale, `labels.ts`.
- F2 pages: Start, Ablauf, Location, Anreise & Übernachtung, Dresscode, Geschenke, FAQ, Kontakt, Datenschutz.
- Hero photo, countdown.
- F11: the accessibility pass — keyboard, 200% zoom, contrast, touch targets, focus rings, screen-reader sweep of the RSVP form.

Blocked on facts we do not control: venue, schedule, dress code, gift wishes. See the table below — these are now the most urgent non-technical items, because send-out is only weeks away rather than months.

A thin Ablauf page is acceptable at launch and can be redeployed. An empty Location page is not — the venue is the first thing a guest looks for.

### M4 — Send-out readiness

- Seed all ~60 households, generate codes, export `codes.csv`.
- **Gate 2:** codes are printed only after F1 has been used end to end on real hardware. A format bug found after 60 cards are printed is unrecoverable.
- **Gate 3:** rehearse a backup restore — `VACUUM INTO`, restore into a fresh container, confirm codes and guest list survive. An untested backup is not a backup, and after send-out the guest list is irreplaceable.
- F6 dashboard, *if the time is there*. Otherwise it moves to M5.

Order in the final week: freeze → seed → export → print → **proof-read a physical card** → send. The proof-read catches what no test does, such as a code that is technically valid and visually ambiguous in the chosen typeface.

### Send-out

Cards in the post. Site live, RSVP open.

In the first fortnight, watch `last_login_at`. A household that never logs in needs a phone call, not a feature.

### M5 — The long middle

This is now the longest stretch of the plan — roughly send-out until mid-May. Everything in it is optional and cuttable.

- F6 dashboard, if it slipped out of M4. Worth doing early in this window: it is the tool you will actually live in for six months.
- F8 budget. No dependencies; bring it forward whenever the mood strikes.
- F9 curated gallery, including the thumbnailing library decision that [04-architecture](04-architecture.md) deferred.
- RSVP chasing off the nudge list. Expect the last 20% of households to need a personal nudge regardless of how good the site is.

### RSVP deadline · ~2 months before the wedding

The form goes read-only. Late changes go through us in the admin UI.

**Gate 4:** seating starts now, not before. Assigning seats against a moving headcount wastes the effort twice.

Still to decide, before this date: what the site shows between the deadline and seating publication — a few weeks in which the site has little to say, and during which guests visit most.

### M6 — Seating

- The hand-drawn church and party floor-plan SVGs, with stable `id` attributes per unit and per seat, checked into the frontend. **Draw them before this milestone starts**, since they depend on the venue's and the church's final layout.
- Seating unit and seat CRUD, per-seat assignment UI, stale-assignment detection.
- Guest view behind `seating_published`.
- Printable output for the day.

Finish roughly three weeks before the wedding, so the venue and caterer get final numbers with room to breathe.

### M7 — Wedding week

- Final headcount to the caterer, on whatever date their contract names.
- Publish seating.
- Print: seating plan for the wall, place cards, the kitchen's allergy sheet.
- Code freeze. No deploys in the final week — a broken site on the day is worse than any missing feature.
- Take a backup, verify it, and stop thinking about the software.

### M8 — After the wedding

Open F10 guest uploads via `uploads_open`: quotas, moderation, ZIP export. Announce it once, wherever the guests actually are.

Gate 5 is satisfied by construction — there is nothing to upload before the wedding.

### M9 — Wind-down

Roughly three months post-wedding, run the end-of-life procedure in [06-privacy-security](06-privacy-security.md): export the archive, `VACUUM INTO` a final snapshot, remove the container, delete the volumes and the stale copies. Photo retention is still open and gates this step.

## Consequences of an early send-out

Posting the invitations nine months before the wedding is unusually early — it merges the save-the-date and the invitation into one mailing. That is a legitimate choice and it removes a whole phase of work, but three consequences follow and are worth planning around:

1. **The RSVP window is about six months long.** Most answers will arrive in the first three weeks and the last few in May. Plan for a long, quiet chasing period rather than a single deadline push.
2. **Answers go stale.** Someone who says yes in November may be pregnant, moved, or separated by May. The form stays editable until the deadline precisely for this, and the deadline is where the numbers finally stop moving — which is also when per-head budget items stop drifting.
3. **The field-set freeze lands very early.** Gate 1 is now weeks away rather than months. That is uncomfortable, but it is also the correct incentive: the F3 story has to be written properly before anything is built, because there is no slack in which to discover a missing field.

One phase drops out entirely: **no separate save-the-date.** A save-the-date solves two problems — an invitation too far out to be actionable, and guests who must book travel far ahead — and neither applies at nine months with a mostly local guest list. It would also mean either two variable-data print runs or a card pointing at a site nobody can log into yet. An Oct/Nov mailing additionally lands exactly when employers collect holiday planning for 2027, which is the save-the-date's real job done by the invitation itself. The one argument for splitting — hedging an unfixed date — disappears once the venue is booked, which happens well before anything is printed.

## Facts needed, in dependency order

All are inputs we do not control. The urgency has changed considerably: several of these are now needed within weeks, not months.

| Fact | Blocks | Urgency |
|---|---|---|
| Venue and church: names, addresses, travel and accommodation notes | M3 | Being fixed by ~early September 2026. Also fixes the date |
| Schedule of the day, dress code, gift wishes, bank details | M3 | **Before send-out** — now the highest-urgency unknown |
| RSVP deadline: the exact date, and what the site says after it | M2 configuration, and the wording printed on the card | **Before send-out** |
| Soft cap on guest-added members | M2 | Before send-out |
| Final guest list and household groupings | M4 | Before send-out — codes are printed per household |
| Caterer age brackets | F6 dashboard | Can slip; derived at read time |
| Reception room and table layout | The SVG, and therefore M6 | Spring 2027 |
| Who draws the SVG, in what tool, ids surviving re-export | M6 | Spring 2027 |
| Photo retention policy | M9 and the Datenschutz page | Late |

The RSVP deadline is worth deciding early even though it is enforced late: it should be **printed on the invitation card**, and a card that says "bitte bis …" is far more effective than a website that says it.

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Development does not finish before send-out | **Medium–high** — the window is short and the must-ship list is most of the app | Cut F6 to a raw CSV export and ship without the dashboard. Never cut F1, F3, F4 |
| Content facts arrive after the code is ready | Medium — venue and date land in early September, which removes most of this | The blocker is content, not software. A thin Ablauf page can be redeployed; the Location page cannot be empty |
| Date changes after the cards are printed | **Low, and now effectively closed** — the venue contract fixes the date well before M4 | Do not print until the venue is booked. That ordering is free and it is the only protection that matters |
| A code format bug found after printing | Low, catastrophic | Gate 2, plus a physically proof-read card before the run |
| Guests answer in November and their plans change by May | Certain, for some | The form stays editable to the deadline; the deadline is where numbers freeze |
| Nobody uses the site and everyone phones instead | Medium | Measured, not guessed: `last_login_at` and the nudge list from week one. The fallback is a phone call per household, which is what would have happened anyway |
| SVG re-export drops the `id` attributes | Medium | A test that fails if any `seating_unit.svg_element_id` or `seat.svg_element_id` has no matching shape. Catches it at build time instead of in June |
| Server dies close to the date | Low | Gate 3. A rehearsed restore turns this into an afternoon |
| Guests answer by WhatsApp instead of the site | High | Not a software problem. Enter it in admin; the dashboard is the source of truth, not the form |

## Deliberately not scheduled

- A staging environment. Deploys are cheap, the audience is 80 people, and a second environment is a second thing to maintain. The integration suite plus a local run is the safety net.
- CI beyond `go test ./...` locally, unless it turns out to be free. Still open in the TODO.
- Any feature not in [02-features](02-features.md). The freeze applies to ambition as much as to columns.
