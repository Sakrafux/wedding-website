# 06 — Privacy & Security

Status: draft · Last updated: 2026-08-22

What personal data this app holds, why it is allowed to hold it, how long, and what protects it. Referenced from [03-data-model](03-data-model.md) and [04-architecture](04-architecture.md).

The governing constraint is the threat model from [01-vision-scope](01-vision-scope.md): guests are friends and family. We defend against guest **mistakes** and drive-by strangers, not against a determined insider. The one hard boundary is admin-only data — budget above all — which a guest session must never reach. Everything below is calibrated to that, and says so where it deliberately stops short.

## What we hold

| Data | Where | Why we need it | Visible to |
|---|---|---|---|
| Household display name | `household.display_name` | Login confirmation, addressing the RSVP | Own household, admin |
| Household login code | `household.code` | The only authentication secret | **Admin only** — never in a guest response |
| Guest first/last name | `guest` | Guest list, seating, place cards | Own household, admin. Tablemates' names only via the seating view, once published |
| Attendance scope | `guest.attending` | Headcounts for church, party, caterer | Own household, admin |
| Meal choice, portion, midnight snack | `guest` | Caterer order | Own household, admin |
| Allergies and intolerances | `guest.dietary_note` | Kitchen safety | Own household, admin |
| Child's age | `guest.age` | Caterer pricing brackets, venue counts | Own household, admin |
| Seating need | `guest.seating_need` | High chairs, wheelchair space | Own household, admin |
| Pram flag | `household.has_stroller` | Floor space next to a table | Own household, admin |
| Free-text household note | `household.rsvp_note` | The catch-all for anything the form missed | Own household, admin |
| Private admin note | `household.admin_note` | Our own planning remarks | **Admin only** |
| Transport seats needed/offered | `household` | Shuttle capacity gap | Own household, admin |
| Seat assignment | `seat_assignment` | The seating plan, church and party | Own household (own unit only), admin |
| Session records incl. IP and user agent | `session` | Session validity and the audit trail | Nobody in the UI; admin via DB |
| Login and change history | `audit_log` | Settle "but I said we were coming" | **Admin only** |
| Photos and their EXIF | `photo` + `PHOTO_DIR` | The gallery | All logged-in guests |
| Request logs (IP, path, status) | stdout → host log collector | Debugging, rate limiting | Server operator |

We deliberately do **not** hold: email addresses, phone numbers, postal addresses, payment data, or any third-party analytics identifier. `household.phone` was considered and rejected in [03-data-model](03-data-model.md).

### The two fields worth a second thought

- **`dietary_note`** will in practice contain a bit of health information — allergies, intolerances, occasionally a pregnancy. Guests volunteer it for an obvious purpose and it goes to the kitchen. The only handling rule worth stating: send the caterer export to the caterer, not into a group chat.
- **`admin_note`** is not sensitive by category but is by nature — it is where a candid remark about a guest would land. DTO-excluded at the type level for exactly that reason. This is the one field where an accidental disclosure would actually cause hurt feelings, which is a better reason to be careful than any statute.

## Legal posture

This is a private, invitation-only site for a family event, run by two private individuals with no commercial purpose. GDPR Article 2(2)(c) — the "purely personal or household activity" exemption — very likely applies. That is a reason for proportionality, not for carelessness: the exemption is narrow, it is contested at the edges, and it does nothing to reduce the actual harm if the data leaks.

So the posture is: **behave as though GDPR applied, at a scale proportionate to 80 guests.**

- **Purpose limitation.** Data is used to plan and run the wedding. Nothing else. No marketing, no sharing with anyone but the vendors who need a headcount.
- **Data minimisation.** Every field in [03-data-model](03-data-model.md) has to justify itself. The rejected-fields table there is part of this document's argument.
- **Transparency.** A short German "Datenschutz" page, reachable without logging in, states in plain language: what we store, why, who sees it, how long, and how to reach us to change or delete it. Written in the same informal register as the rest of the site — a wall of legalese would be worse than nothing here.
- **Rights.** Access, correction and deletion are handled informally: a guest asks us, we do it in the admin UI. At 60 households a formal process would be theatre.
- **No Impressum.** German TMG/DDG imprint duty attaches to business-like ("geschäftsmäßig") telemedia. A private wedding invitation behind a login is not that. Contact details are on the site anyway, on the Kontakt page.
- **No processors.** No SaaS in the critical path, no CDN, no analytics, no Google Fonts, no error-reporting service. That is a privacy decision as much as an architectural one: there is no third party to conclude a DPA with because there is no third party.

**Vendors are the real disclosure surface.** The caterer gets meal counts, portions, allergies and child brackets; the venue gets a headcount and special-seating needs. Export the minimum: for the caterer, names are usually unnecessary — counts plus an allergy list keyed to a table is enough. The CSV export should be reviewed before it is sent, every time.

## Retention and deletion

The site lives from invitation send-out until roughly **three months after the wedding** (from [01-vision-scope](01-vision-scope.md)); gallery lifetime beyond that is still open, see below.

End-of-life procedure, to be run deliberately rather than by letting the server rot:

1. Export what we want to keep as flat files: guest list CSV, final seating plan, budget CSV, and the photo archive.
2. `VACUUM INTO` a final database snapshot; keep it encrypted, offline, as the archive of record.
3. Stop and remove the container, then delete the volumes — `DB_PATH` and `PHOTO_DIR`.
4. Delete stale copies: the print-shop code CSV, caterer exports, backup snapshots on the server, anything that ended up in a downloads folder.

Interim retention:

| Data | Retention |
|---|---|
| Guest and RSVP data | Life of the site, then the encrypted archive snapshot |
| Sessions | Deleted at expiry; expired rows purged at startup and daily |
| Audit log | Life of the site. Not pruned — it is small and its whole value is history |
| Request logs | Whatever the host collector is configured for; keep it short (days, not months) |
| Photos | **Open question** — see below |
| Backups | Rotate; keep at most a few weeks of snapshots, and delete them at end of life with everything else |

A guest who asks to be removed before the wedding gets their row soft-deleted, which drops them from every count. Hard deletion happens at end of life for everyone at once.

## Handling the database file

`household.code` is stored in **plaintext**. Deliberate trade-off, recorded in [03-data-model](03-data-model.md): a code must be readable back to a guest who lost their card, and it is the sole authentication factor, so hashing would only protect against someone who already has the file.

What that actually means, in proportion: the file contains a guest list, meal preferences, and the codes that unlock a wedding invitation. It sits on a personal server. There is no financial data, no credentials reused anywhere else, and nothing a stranger would want. So the handling rules are ordinary hygiene, not a key-management regime:

- Don't commit a database, a dump, or a `.env` to the repository. Nothing that looks like a real code goes into a spec, an issue, or a screenshot.
- Keep backups on the server or somewhere private. Encrypt them if they go somewhere shared; a cloud drive that only you can read is fine.
- `ADMIN_PASSWORD` is plaintext in the environment. Readable only with server access, which already implies access to the database, so the environment variable is not the weak link. Still make it long and random rather than memorable — it is typed rarely.
- The print-shop codes CSV is the one copy that leaves your control. Send it, get the cards, delete it on both ends. Not because someone would abuse it, but because a stray file nobody owns is how these things get forgotten in a downloads folder for five years.

If the file did get out, the response scales with when: before printing, regenerate every code at zero cost; after send-out, regenerating invalidates every printed card, so realistically you shrug, rotate the admin password, and `DELETE FROM session`. The audit log is there if you want to see whether anything actually happened.

## Authentication

### Code strength

Alphabet of 32 characters, length 6 → 32⁶ ≈ **1.07 billion** possible codes. With ~60 in use, a single blind guess hits a valid code with probability ≈ 5.6 × 10⁻⁸. Against the rate limit below, an attacker gets on the order of 10 tries an hour per IP, so the expected time to a first hit is measured in centuries per IP. Online guessing is not a threat; the offline file is.

Codes are generated with `crypto/rand`, never `math/rand`, and rejected on collision with an existing code (retry, do not increment).

### Rate limiting

Per-IP limiting on both login endpoints, on the order of **10 failures per hour**, with a friendly German message and **never a lockout** — locking out a confused 75-year-old is a worse outcome than the attack it prevents. Successful logins do not consume budget.

- Client IP comes from `X-Forwarded-For`, trusted **only** when the request arrives from a CIDR in `TRUSTED_PROXY_CIDRS`. Otherwise the header is attacker-controlled and the limit is bypassable by anyone who reads this document.
- A second, global failure counter across all IPs bounds distributed guessing. Crossing it is a signal to look at the logs, not a reason to lock anyone out.
- Every failure is written to `audit_log` as `login_failed`, with IP and user agent. **The attempted code is not logged** — a near-miss log would be a partial key list, and a typo log would contain another household's real code.
- The admin login endpoint gets a stricter limit than the guest one. It is the only door where guessing is worth an attacker's time.

### Sessions

- Token: 256 bits from `crypto/rand`, base64url-encoded.
- Stored as a **SHA-256 hash** in `session.id`; the raw token exists only in the cookie. A leaked database therefore does not hand over live sessions — though it does hand over the codes, so this is defence in depth, not a barrier.
- Cookie: `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`. The site has its own subdomain, so a root-scoped cookie reaches nothing but this app; the earlier path-scoped variant existed only while it shared a hostname. `SESSION_COOKIE_SECURE=false` exists **only** for local development over plain HTTP and must never be false in production.
- Household sessions: 365 days, rolling refresh on use (extend at most once a day to avoid a write per request). Long life is a deliberate UX choice — "log in once, ever" is a stated product goal, and the alternative is a phone call.
- Admin sessions: hours, no rolling refresh. Different risk profile, as recorded in [04-architecture](04-architecture.md).
- Logout deletes the session row, so revocation is immediate and real. This is why sessions are opaque DB tokens rather than JWTs.
- Session lookup compares hashes, and the token is never logged, never in a URL, never in a query parameter.
- Expired rows are purged at startup and once a day.

### Authorisation

- Every mutating request re-checks that the session's household owns the target row, server-side. Ownership is never inferred from an ID the frontend sent.
- `/api/admin/*` is rejected wholesale for household sessions by middleware, before any handler runs. Budget endpoints exist only under that prefix — there is no budget field anywhere else in the API surface.
- A hidden nav link is not a security control and is never counted as one.
- Guests see their **own** table's occupants in the seating view, not the whole plan's names. That is a privacy decision, not just a UI one.

## Response hygiene

**DTOs are a privacy control.** Domain structs are never serialised. Every guest-facing response is an explicit type in `web/dto` that structurally cannot contain `code`, `admin_note`, or budget data — so adding a column to `household` cannot silently start leaking a login code. This is the single most valuable rule in [04-architecture](04-architecture.md), and integration tests assert it by scanning guest responses for those field names.

Errors return `{ "error": { "code": "...", "message": "..." } }`. The German message is safe to show a guest verbatim; stack traces, SQL, and file paths never reach the client. A request ID appears in both the log and the message so a guest can read it out over the phone.

Login responses do not distinguish "no such code" from any other failure, and take a comparable amount of time either way.

## Transport and browser hardening

TLS terminates at the existing reverse proxy; the Go process speaks plain HTTP on the internal network and manages no certificates. HSTS is the proxy's job.

Headers set by the app on every response:

| Header | Value | Why |
|---|---|---|
| `Content-Security-Policy` | `default-src 'self'; img-src 'self' data:; style-src 'self'; font-src 'self'; script-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'` | Everything is same-origin already — no CDN, no external fonts — so a strict policy costs nothing |
| `X-Content-Type-Options` | `nosniff` | Stops a photo being interpreted as script |
| `Referrer-Policy` | `no-referrer` | Nothing external to leak a referrer to, and nothing to gain from sending one |
| `X-Frame-Options` | `DENY` | Belt and braces with `frame-ancestors` |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` | No feature here needs them |
| `X-Robots-Tag` | `noindex, nofollow` | Plus a `robots.txt` disallow. "Nothing is indexed" is a stated principle |

**CSRF:** `SameSite=Lax` blocks cross-site POST with cookies, the API is same-origin, and all mutations are `POST`/`PUT`/`DELETE` with a JSON body — never a `GET` with side effects. Handlers additionally require `Content-Type: application/json`, which a cross-site HTML form cannot send. For this threat model that is sufficient; a synchroniser token is recorded as rejected below.

**XSS:** React escapes by default and `dangerouslySetInnerHTML` is banned outright. Guest notes, captions and names are rendered as text. Page content is hardcoded in components — there is no CMS, so there is no stored-content injection path.

**SQL injection:** sqlx with bound parameters everywhere. No string-built SQL, including in admin filters and CSV exports.

## Photos

**The governing analogy: this is a shared photo album among wedding guests.** It carries exactly the responsibility that sending the same photos into a family WhatsApp group would — which is to say, essentially none on our side. A guest who uploads a photo is sharing it with the other guests, knowingly, the same way they would anywhere else. We host it and gate it behind a login; we do not act as a custodian of what is in it.

That settles the EXIF question, which [03-data-model](03-data-model.md) already decided on archive-quality grounds: **originals keep their metadata, GPS included.** Stripping it would break orientation tags, degrade the archive, and buy protection against a scenario the threat model does not contain. Thumbnails and lightbox images happen to lose metadata anyway, since they are re-encoded — a side effect, not a control. If a photo is ever published outside this site, strip GPS at that point.

The remaining rules are cheap technical hygiene, not privacy controls:

- Uploads validated by **content sniffing**, never by file extension: JPEG, PNG, HEIC, MP4. **SVG rejected** — an SVG is a script container, and serving one from our own origin is the one upload that could actually break something.
- Content-addressed storage names. The user-supplied filename is display metadata and never a path component — no traversal, no overwrite.
- Served by an authenticated handler with an explicit `Content-Type` and `nosniff`. No unauthenticated direct-file URL, so the gallery does not become a public image host.
- Per-request size cap and a per-household quota, to bound a mis-tap that uploads a 400-photo camera roll. Quota values still open.
- Guests can delete their own uploads; admins can hide (row and file kept, not served) or delete outright. That, plus asking each other, is the whole moderation story.

## Logging and the audit log

`httplog` writes structured request logs to stdout: method, path, status, duration, request ID, client IP, user agent.

Explicitly never logged: login codes, session tokens, cookie headers, `Authorization` (there is none), and request bodies on the two auth endpoints. Guest names in logs are incidental and acceptable; codes in logs would be a key leak.

The **audit log** exists to answer disputes — "but I said we were coming", "who deleted Tante Erna?" — and to make an unexpected login visible after the fact. It is append-only, admin-only, and stores JSON snapshots of changed fields only. It is not a surveillance tool and is not shown to guests. Because it contains the same personal data as the tables it shadows, it is covered by the same retention and deletion rules.

## Operational

- Container runs as a **non-root user**; the runtime image is distroless or scratch, which the pure-Go SQLite driver makes possible. Smaller image, smaller attack surface, no shell to land in.
- Only the app port is published, and only to the reverse proxy — the container is not reachable from the internet directly.
- Volumes are the only writable paths.
- Dependencies are few and boring by design. `go mod tidy` plus `govulncheck`, and `pnpm audit` for the frontend, before each deploy is proportionate; a continuous scanning setup is not. Frontend versions are fixed by `pnpm-lock.yaml` and installed with `--frozen-lockfile`; pnpm's default refusal to run install scripts stays on, since an unreviewed `postinstall` is the realistic supply-chain risk, not a patch-level drift.
- **Rehearse a backup restore before invitations go out** — gate 3 in the TODO. Also verify that the restored database actually contains the codes, because after send-out the guest list is irreplaceable.
- Never `cp` a live WAL-mode SQLite file; use `VACUUM INTO`. A torn backup is worse than none, because it looks like one.

## What we deliberately do not do

Stated so that the absences read as decisions rather than oversights.

| Not doing | Why |
|---|---|
| Hashing login codes | Must be readable back for a guest who lost the card; it is the sole factor, so hashing only protects against an attacker who already has the file. Recorded as accepted risk. |
| Two-factor authentication for guests | We have no second channel — no verified emails, no phone numbers — and it would break the "log in once, ever" goal for the exact people it would harm most. |
| Bcrypt/Argon2 for the admin password | Single admin, plaintext env var by design. Nothing to hash against. |
| Encryption at rest of the database | Key would have to live on the same server as the file. Real protection comes from server access control and encrypted **off-server** backups. |
| CSRF synchroniser tokens | `SameSite=Lax` + same-origin + JSON-only mutations covers the realistic attack at this threat level. |
| Account lockout after failed logins | Worse than the attack: it turns a typo into a phone call, and hands anyone a trivial denial of service against a specific household. |
| Consent banner / cookie notice | One strictly necessary session cookie, no analytics, no third parties. A banner would be noise. |
| Analytics, error reporting, uptime SaaS | Every one is a third-party data flow and a DPA we do not want. Logs and the audit table are enough at this scale. |
| Formal DPIA, ROPA, DPA paperwork | Disproportionate for a private household-activity site with no processors. |
| Defending against a malicious logged-in guest | Out of scope by the stated threat model — except for admin-only data, which is enforced server-side and is the one place the exception does not apply. |

## Open

- **Photo retention** after the site goes offline: how long the gallery stays up, and whether the archive is handed to guests (a download link, a USB stick) or kept privately. Blocks the end-of-life procedure above.
- **Upload quotas** — proposed 100 files / 2 GB per household, still to confirm.
- Exact wording of the German Datenschutz page. Should be drafted alongside the F2 content pages.
- Whether the caterer export can be name-free (counts and allergies keyed by table) rather than a full guest list. Would remove the largest planned disclosure.
