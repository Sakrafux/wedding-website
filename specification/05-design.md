# 05 — Design System

Status: draft · Last updated: 2026-10-11

This is the visual and interaction contract for the frontend. It exists so that components are built from a fixed vocabulary instead of ad-hoc Tailwind classes, and so that accessibility is a property of the system rather than something remembered per screen.

Non-negotiable premise, restated from [01-vision-scope](01-vision-scope.md): a 75-year-old must be able to RSVP alone, on a phone, without calling us. Every decision below is downstream of that.

## Direction

**Festive and traditional, on white.** Pure white paper — the printed cards are white — with cream reserved for cards and input fills, serif headings in Cormorant 700, generous whitespace, a berry primary and honey-gold attention strokes. The site should feel like the printed invitation, not like a SaaS dashboard — for guests. The admin area shares the tokens but drops the decorative layer in favour of density.

Decided inputs:

- The invitation cards exist and are clean white; the site's paper matches them. The first draft had a cream page, which read as yellowed next to the real cards (reworked 2026-10-11, after previewing the alternatives with the couple).
- No stationery colours or fonts to match beyond that. If stationery is designed later, it copies from here.
- Engagement photography **is** available, so the start page has a real hero image from day one.
- **No dark mode.** Rejected deliberately — see below.
- Informal **"du"** throughout the German copy.

## Colour

Warm neutrals plus two accents. Two accents rather than one because the app has two distinct kinds of emphasis: *interactive* (buttons, links, focus) and *attention* (your own table, unread notes, warnings). Collapsing them into one colour makes a warning look like a button.

### Tokens

| Token | Hex | Role |
|---|---|---|
| `paper` | `#FFFFFF` | Page background. Pure white, like the printed cards. |
| `surface` | `#FAF7F0` | Cards, sheets, inputs — cream, the lift off the white page. |
| `surface-sunken` | `#F3EEE4` | Table stripes, disabled fields, read-only RSVP after deadline. |
| `ink` | `#332F2D` | Body text. Warm near-black, never `#000`. |
| `ink-muted` | `#6E6862` | Secondary text, labels, help text. |
| `line` | `#E4DCCB` | Borders, dividers, input outlines. |
| `primary` | `#8A4254` | Berry. All interactive elements: buttons, links, focus rings, active nav. |
| `primary-hover` | `#733647` | Hover/pressed state. |
| `primary-soft` | `#F8E9EE` | Pastell rose tint: selected option, own-table fill in the seating SVG. |
| `accent` | `#A87D15` | Honey gold. Attention strokes only — own-table outline, alert border, unread-note border — never a fill under text. |
| `accent-strong` | `#6D540F` | Bronze — gold as **text** at any size. |
| `accent-soft` | `#F8F1DC` | Champagne tint: attention backgrounds. |
| `success` | `#3F6B4A` | RSVP saved confirmation. |
| `warning` | `#A2570F` | Over-capacity table, addition limit reached. Admin-facing mostly. Orange, not amber — gold owns the yellow hue now. |
| `danger` | `#9A3324` | Destructive confirmation, validation errors. |

### Measured contrast on `paper` (#FFFFFF)

| Pair | Ratio | Verdict |
|---|---|---|
| `ink` on `paper` | 13.3:1 | AAA body text |
| `ink-muted` on `paper` | 5.5:1 | AA body text; not for anything below 16px |
| `ink-muted` on `surface-sunken` | 4.75:1 | Read-only RSVP text after the deadline — the tightest pair in the system |
| `primary` on `paper` | 7.0:1 | AA at any size; also the button-fill/white-text ratio |
| `accent` on `paper` | 3.7:1 | UI boundaries and large text only — never body text, never a fill under text |
| `accent` on `primary-soft` | 3.2:1 | The seating-plan own-table stroke; still ≥3:1 as a UI boundary |
| `accent-strong` on `paper` | 7.2:1 | AA at any size — use this whenever gold is text |
| `accent-strong` on `accent-soft` | 6.4:1 | Badge text on the champagne tint |

Rule: **`accent` is a stroke colour, `accent-strong` is a text colour.** A yellow that clears 4.5:1 on white is forced into brown ochre — the reason gold is honey for strokes (the ≥3:1 UI-boundary rule) and bronze for text. Nothing may put body text on a gold fill.

### Colour is never the only signal

Guest-added members, unread notes, stale seat assignments, and "your table" each carry an icon or a text label in addition to their colour. Roughly 8% of men are red-green colour blind; under simulated deuteranopia berry collapses to grey while gold, danger and warning all drift into olives of differing lightness (Viénot/Brettel simulation). Hue does most of the primary/accent work, lightness carries the status colours, and the non-colour signal carries whatever remains.

## Typography

Self-hosted `.woff2`, subset to Latin, bundled into the Vite build. **No Google Fonts CDN** — it is an external dependency in the critical path and it leaks guest IPs to a third party, both of which the specs forbid.

| Role | Face | Notes |
|---|---|---|
| Display | **Cormorant Garamond**, 700 | Names, page titles, the hero. High contrast, only at large sizes. 700 since 2026-10-11 — 600 read as too thin next to the white paper; the 600 files left the bundle because nothing else requested them. |
| Body & UI | **Source Serif 4**, 400/600 | Designed for screen text; stays warm and traditional while remaining legible at 18px on a cheap Android panel. |
| Numeric | Source Serif 4, tabular figures | Budget tables, headcounts, seat numbers. `font-variant-numeric: tabular-nums`. |

Rejected: body text in Cormorant (too thin and too high-contrast at paragraph size — it looks elegant on a mockup and hurts on a phone); a sans body such as Inter (legible, but breaks the printed-invitation feel that the whole direction rests on); EB Garamond for body (closer than Cormorant, still lighter than Source Serif at 18px).

### Scale

Base **18px**, not the 16px web default. `rem`-based throughout, so browser zoom and OS font scaling work. Line height 1.6 for body, 1.2 for display.

| Step | Size | Use |
|---|---|---|
| `display` | 40 / 56px | Hero names only |
| `h1` | 30 / 36px | Page title |
| `h2` | 24 / 28px | Section |
| `h3` | 20 / 22px | Sub-section, card title |
| `body` | 18px | Default |
| `small` | 16px | Help text, captions. **Floor — nothing smaller ships.** |

Two values = mobile / desktop. Mobile is the design target; desktop is the adaptation.

Line length capped at ~66 characters (`max-w-prose`) on informational pages.

## Layout & spacing

Spacing scale (Tailwind default 4px base): `2 4 6 8 12 16 24 32 48 64`. Nothing off-scale.

- Single column on mobile. Content width `max-w-2xl` centred; admin tables get `max-w-6xl`.
- Page gutter 16px mobile / 24px desktop.
- Vertical rhythm between sections: 48px mobile, 64px desktop.
- Border radius: 8px for inputs and buttons, 12px for cards. Consistent, never mixed.
- Elevation: one soft shadow (`0 1px 3px rgb(58 50 38 / 0.08)`) for cards. No layered shadow system; the paper metaphor does the work.

**Touch targets are minimum 48×48px** with at least 8px between adjacent targets. This is the single most important accessibility rule here — most RSVP failures for this audience will be mis-taps, not misunderstandings.

## Navigation

Mobile: a fixed bottom bar with 4–5 items, icon + German label (labels always visible, never icon-only). Overflow items live on a "Mehr" page rather than a hamburger, because a hamburger hides exactly the things an unconfident guest is looking for.

Desktop: horizontal top nav.

The RSVP entry point is persistent and visually primary until the household has submitted; afterwards it becomes "Antwort ändern".

## Component inventory

Built on shadcn/ui (Radix primitives, copied into the repo).

**This list is a planning draft, not a contract.** It is neither exhaustive nor binding: it exists to show that the feature set decomposes sensibly and to catch missing pieces early. Implementation may add, merge, split or drop components freely as the real screens take shape. What *is* binding is everything above it — tokens, type scale, spacing, touch targets, and the accessibility rules — because those are the parts that are expensive to retrofit.

**Primitives** — `Button` (primary/secondary/ghost/danger), `Input`, `Textarea`, `Select`, `RadioCardGroup`, `Checkbox`, `Switch`, `Label`, `FieldError`, `Badge`, `Card`, `Alert`, `Dialog`, `Sheet`, `Skeleton`, `Toast`.

**Composed, guest-facing:**

| Component | Purpose |
|---|---|
| `CodeInput` | Login. Big monospaced-feel field, auto-uppercase, strips whitespace and dashes as you type, so the value in the field is the value that gets sent. The card prints the code ungrouped (`ABC234`), which is what removed the dash problem rather than solving it. `inputmode="text"` `autocapitalize="characters"`, no autocorrect. The hint sits outside the field, never as a placeholder. |
| `HouseholdConfirmCard` | "Willkommen, Familie Müller — bist das du?" with a clear "Nein, zurück". |
| `HeroSection` | Photo, names, date, countdown. |
| `CountdownBadge` | Days until the wedding. Hidden after the day. |
| `ScheduleTimeline` | Ablauf as a vertical timeline. |
| `InfoSection` | Generic heading + prose block for hardcoded content pages. |
| `RsvpMemberCard` | One card per member: attendance scope, then the catering fields **revealed only** for `party_only`/`both`. |
| `ScopeSelector` | The household-level "Wir kommen zu:" bulk setter. Large radio cards, not a `<select>`. |
| `AddPlusOneSheet` | One field: the companion's name. Offered only to a household of one that has not already added somebody; every other household sees an explanation and our phone number in its place, never a disabled button. |
| `TransportFields` | Seats needed / seats offered steppers, with plain-language explanation of what they are for. |
| `RsvpSummary` | Post-submit recap of exactly what was saved. Unmissable. |
| `SeatingMap` | Renders the checked-in floor-plan SVG; colours the own table `primary-soft` with an `accent` outline, labels it with names. Pan/zoom on mobile. |
| `GalleryGrid` / `Lightbox` | Curated gallery. |
| `UploadDropzone` | Post-wedding uploads, with per-file progress and retry. |

**Composed, admin-facing:** `StatTile`, `HeadcountPanel`, `DietaryList`, `DeltaList`, `NoteInbox`, `NudgeList`, `TableEditor`, `AssignmentPicker`, `StaleAssignmentAlert`, `BudgetTable`, `BudgetRollup`, `PhotoModerationGrid`, `CsvExportButton`.

Admin uses the same tokens with denser spacing and the decorative layer (hero, serif display sizes) dropped.

### Form behaviour

- Every input has a visible `<label>`. Placeholders are never labels.
- **Every guest-facing field carries a help text, behind a help button.** A small `?` icon button sits beside the label and opens a popover with one or two sentences: what the field is for, and what answer we expect. Not inline under the field — "Mitternachtssnack", "Plätze angeboten" and "Alter am Hochzeitstag" all need explaining and all of them explained at once turns the RSVP form into a wall of grey text that gets skipped, which costs the guests who actually needed it. The popover also removes the pressure to write help that fits in half a line. Rules: the button is a real `<button>` with an accessible name naming its field ("Hilfe zu Mitternachtssnack"), not a bare "?"; it meets the 48×48 target with an invisible hit area if the icon is smaller; it is a Radix `Popover`, so Escape and click-outside close it and focus returns; it is never the only route to the information for anything a guest *must* understand to answer correctly — a field whose label cannot be understood without the popover has the wrong label. `portion: none` keeps its inline hint, since it disambiguates two options rather than explaining one field. Admin fields are exempt. Added 2026-08-31, after the F3/F4 stories were written — retrofit where they are silent about it.
- Errors appear **under the field**, in `danger`, with an icon, and the field gets `aria-invalid`. A summary at the top of the form lists them and links to each field.
- Validation on blur and on submit; never on every keystroke.
- The RSVP form autosaves nothing. There is one explicit "Speichern" button, because a silent autosave gives an unconfident guest no moment of completion.
- After the RSVP deadline the form renders read-only on `surface-sunken` with an explanatory `Alert`, not disabled inputs — disabled text fails contrast and reads as broken.
- **An editable field is filled `paper`, never transparent.** Added 2026-09-01 after `F11-05`, re-pointed 2026-10-11 when paper and surface swapped roles: the white fill is what makes an editable field read as a hole in the cream card, and a transparent one would be cream on cream — exactly what a disabled control looks like. The shadcn defaults arrived transparent and had to be corrected. `Input`, `Textarea` and `Select` share one fill, one radius (`rounded-lg`) and one type size; three controls in one form looking like three kinds of control is the other half of the same complaint. The `md:text-sm` step shadcn ships with is removed: 16px is the floor for this audience, and a desktop breakpoint must not go under it.
- **A household that has already answered sees its answer, not a filled-in form.** `F3-F09`, 2026-09-01. The summary is a view of the stored answer with "Antwort ändern" beside it, and a save returns to it. As a post-save-only screen it could never be found again, which left a household re-reading the form to work out whether their last edit had gone in. The admin form is form-first — we open it mid-phone-call to change something.
- **A confirmation carries the time and stays.** "Zuletzt gespeichert um 14:32", not "Gespeichert." (`F5-F05`). On a local admin screen the request finishes fast enough that a transient chip is a flash; a confirmation that persists needs no artificial minimum duration, which was the alternative and would have slowed down the screen used sixty times in a row.
- **A pending state may not replace a page a guest is already looking at.** `F11-04`, 2026-09-01: the full-screen skeleton belongs to the cold load, and a navigation between two pages either renders the next page or shows a content-shaped pending state inside the navigation. A guard that awaits a value it already has produces the first kind for free — and it looked like the whole app reloading on every tap.

## Accessibility rules

Beyond the token-level guarantees:

1. Contrast ≥ 4.5:1 for all text, ≥ 3:1 for UI boundaries and icons. Values above are measured, not estimated.
2. Visible focus ring on **every** interactive element: 2px `primary` outline, 2px offset. Never `outline: none` without a replacement.
3. Fully usable at 200% browser zoom and at OS font scale 1.5× — no horizontal scrolling, no clipped text.
4. Semantic HTML first, ARIA only where semantics fall short. Radix handles the hard cases.
5. Skip-to-content link, and one `<h1>` per page with no skipped heading levels.
6. `prefers-reduced-motion` respected; all motion is optional decoration.
7. Every image has meaningful `alt`; decorative images get `alt=""`.
8. The seating SVG is not the only route to the answer — the guest's table and tablemates are also rendered as plain text beneath it. An SVG floor plan is unusable with a screen reader or at high zoom.
9. Language declared `lang="de"`.
10. Error messages name the fix, not the fault: "Der Code besteht aus 6 Zeichen — bitte prüf ihn noch mal", never "Ungültige Eingabe".

## Motion

Minimal. 150ms ease-out for hover and focus, 200ms for sheets and dialogs. No page transitions, no scroll-triggered animation, no parallax. The countdown does not tick a live seconds counter.

## Imagery

- Hero: engagement photography, `object-cover`, `aspect-[4/5]` on mobile and `aspect-[16/9]` on desktop, with a warm overlay so display text keeps its contrast.
- Served as `.webp` with a `.jpg` fallback, explicit `width`/`height` to prevent layout shift, `loading="lazy"` below the fold.
- Gallery thumbnails are server-generated; originals are only fetched by the lightbox.

## German label map

The one and only mapping from English enum values to German UI text. Lives at `web/src/lib/labels.ts` as `const` objects keyed by the enum type. **No German string is written inline in a component.** This is what makes "enums are English everywhere" survivable.

### `guest.attending`

| Value | Label | Short (tables) |
|---|---|---|
| `no` | Kommt nicht | Nein |
| `church_only` | Nur zur Kirche | Kirche |
| `party_only` | Nur zur Feier | Feier |
| `both` | Kirche und Feier | Beides |
| `null` | Noch keine Antwort | Offen |

### `guest.meal_choice`

| Value | Label |
|---|---|
| `all` | Isst alles |
| `vegetarian` | Vegetarisch |
| `vegan` | Vegan |

### `guest.portion`

| Value | Label | Help text |
|---|---|---|
| `none` | Kein Essen | Für Babys oder wenn jemand nicht mitisst |
| `kids` | Kinderportion | |
| `full` | Normale Portion | |

### `guest.seating_need`

| Value | Label |
|---|---|
| `normal` | Normaler Platz |
| `with_parent` | Sitzt bei den Eltern (kein eigener Platz) |
| `high_chair` | Hochstuhl |
| `wheelchair` | Platz für Rollstuhl |

### `guest.kind`

`adult` → Erwachsene:r · `child` → Kind

### `guest.origin`

`seeded` → (no label; the normal case) · `guest_added` → Selbst hinzugefügt *(admin-only)*

### `budget_item.status` *(admin-only)*

`planned` → Geplant · `booked` → Gebucht · `partially_paid` → Teilweise bezahlt · `paid` → Bezahlt · `cancelled` → Storniert

### `audit_log.action` *(admin-only)*

`create` → Angelegt · `update` → Geändert · `delete` → Gelöscht · `login` → Anmeldung · `login_failed` → Fehlgeschlagene Anmeldung

### Tone rules for German copy

- Informal **du** throughout, including error messages. Never mix registers on one screen.
- Prefer neutral phrasing where it reads naturally ("Bitte bis zum 1. Mai zurückmelden") so the register is less conspicuous to older relatives, but do not contort a sentence to avoid `du`.
- Gender-neutral where it costs nothing (`Gäste`, `Personen`); no forced constructions in decorative copy.
- Short sentences. No jargon: never "RSVP", "Session", "Account", "Login-Code" → use "Zusagen", "angemeldet", "dein Code".
- The Gedankenstrich is an **en-dash with spaces** (` – `), never an em-dash — German typographic convention. `labels.ts` was normalized to this on 2026-10-11.
- Numbers as digits (`2 Personen`), dates written out (`17. Juli 2027`).

## Rejected

| Option | Why not |
|---|---|
| Dark mode | Doubles every contrast pair and all QA, for an audience mostly on default-light phones. The cream-paper identity does not translate to dark anyway. |
| Google Fonts CDN | External dependency in the critical path; leaks guest IPs. Self-hosted `.woff2` instead. |
| Cormorant as body text | Too light and too high-contrast at 18px on low-DPI screens. |
| 16px base font | Standard, but this audience should not have to discover browser zoom. |
| Hamburger menu on mobile | Hides the exact links an unconfident guest is hunting for. Bottom bar + "Mehr" page instead. |
| Icon-only navigation | Icon literacy is not universal in this age range. Labels always visible. |
| A single accent colour | Cannot distinguish "press this" from "look at this" — a warning would read as a button. |
| Cream page background (the first draft) | The real invitation cards are clean white, and cream next to them read as yellowed — plus cream with muted earth tones is the default look every AI tool ships. Cream survives only as `surface`, the card tint. |
| Gold dark enough to clear 4.5:1 as a fill with text on it | A yellow at that contrast on white is forced into brown ochre — "schaut wie braun aus", in the fiancée's words. Honey gold for strokes (≥3:1 UI rule) and bronze text instead; see the accent rule above. |
| Pastell primary (light rose buttons) | Cannot clear 4.5:1 on white at any size this audience reads. The pastel lives in the tints — `primary-soft`, stripes, lines — while the interactive colour stays deep berry. |
| Autosaving RSVP form | Removes the moment of completion; guests would not know whether they had answered. |
| Multi-step RSVP wizard | Considered for simplicity. Rejected: it hides the total length, and going back to fix one answer becomes a chore. One long form with clear sections instead. |
| Scroll animations / parallax | Motion cost with no comprehension benefit, and a reduced-motion liability. |
| Component library with its own styling system (Mantine, MUI) | Fights Tailwind; already rejected in [04-architecture](04-architecture.md). |

## Open

- Whether the seating SVG needs a second, phone-specific variant — deferred to the F7 story.
