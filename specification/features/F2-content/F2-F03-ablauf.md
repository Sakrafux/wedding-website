# `F2-F03` — Ablauf

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F01`

## Story

As a guest, I want the schedule of the day as a list of times and what happens at each, so that I know when to be where without asking anybody.

## Scope

**In:**

- `/ablauf`.
- `ScheduleTimeline`: a vertical list of time + title + one line of detail.
- The schedule content, hardcoded as data in the component's module.

**Out:**

- Addresses and travel between the two venues → `F2-F04`. The timeline names the place; the Location page says how to get there.
- Dress code → `F2-F05`.

## Instructions

1. Content is hardcoded, per [02-features](../../02-features.md) — but as an **array of entries** in the module, rendered by one component, not as hand-written markup per row. Twelve copies of the same three-element block is where a typo hides.
2. Times are plain text in the data (`"14:30"`), not `Date` values. There is one timezone, the day is fixed, and parsing a time only to format it back is work that can go wrong at a DST boundary for no gain.
3. Vertical timeline, single column, on mobile and desktop alike. A two-sided alternating timeline is decorative and halves the line length on the device most guests use.
4. Each entry: time, what happens, and at most one line of detail. Anything longer belongs on its own page — a schedule that has to be read is not a schedule.
5. The German strings live in `labels.ts` with the rest of the copy, as one `scheduleLabels` export holding the heading and the entries. It is content, but it is also the German text of the site, and the rule has no exception clause.
6. Times use tabular figures so the column lines up; the token is already set.
7. Mark the church and the reception entries so a guest can see at a glance which parts they are invited to under their scope. Do **not** filter the timeline by the household's answer — a `church_only` guest still wants to know that a party happens, and hiding it would read as a mistake rather than as tact.
8. **Open input, tracked in [TODO.md](../../../TODO.md):** the schedule of the day itself. [07-roadmap](../../07-roadmap.md) accepts a thin Ablauf page at launch, redeployed later, so build the page with what is known — ceremony, reception, dinner, party — and leave detail out rather than inventing times.
9. If a time is genuinely not fixed yet, write the entry without one rather than with a guessed one. A time on this page will be believed.

## Test plan

- [ ] Component: `renderApp("/ablauf")` renders every entry in the data, in order.
- [ ] Component: an entry with no time renders without an empty gap or a stray dash.
- [ ] Component: the church and reception markers appear on the right entries.
- [ ] Accessibility: the timeline is an ordered list in the markup, each entry has its time associated with its heading, and one `<h1>` names the page.

## Done when

- [ ] A guest can read the whole day, in order, on one phone screen's worth of scrolling.
- [ ] Checkbox ticked in `README.md`.
