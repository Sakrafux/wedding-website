# `F2-F09` — Location: an overview, and a page per venue

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F04`

## Story

As a guest, I want the church and the party to have their own pages, so that I am not reading two addresses, two car parks and two arrival notes in one scroll and mixing them up.

## Scope

**In:**

- `/location` keeps a short block per venue plus the transfer section, and links onwards.
- `/location/kirche`: address, map link, parking, how to get there.
- `/location/feier`: the same, plus Übernachtung.

**Out:**

- The venue facts themselves. `F2-F04` stays open: this story splits the page, and the addresses are still "steht noch nicht fest" (TODO.md). Neither page may go out to guests in that state.
- A shuttle timetable. Whether there is a shuttle at all depends on the RSVP answers (`F3-F07`).

## Instructions

1. Übernachtung lives on the party page, not the overview: guests looking for a bed are looking for a bed near where the evening ends, and hotels beside the church would be the wrong list.
2. The transfer section stays on the overview. It belongs to neither venue — it is the bit between them — and it is what a guest reads to decide whether to answer the transport question.
3. `VenueBlock` splits in two: the short teaser on the overview (name, one line, link) and the detail block on the venue page. One component rendering both would grow a `variant` prop for two callers.
4. The detail pages keep the map link as a plain external link, marked as external in the text as well as by the icon. No iframe, for the reasons `F2-F04` gives: a `frame-src` hole in the CSP and every guest's IP handed to a mapping company.
5. Nav: `/location` stays the bar entry. The venue pages are reached from it, not from the bar — five bar entries is the ceiling, and a guest who wants the church address is already on the Location page.
6. `locationLabels` splits along the same lines, with the placeholders staying honest sentences rather than lorem ipsum.

## Test plan

- [ ] Component: `/location` renders both teasers and links to both pages.
- [ ] Component: `/location/kirche` renders the church heading and no accommodation section.
- [ ] Component: `/location/feier` renders the accommodation section.
- [ ] Component: the map link, when a URL exists, is external and says so.

## Done when

- [ ] Each venue's facts are on one page with nothing else's on it.
- [ ] Checkbox ticked in `README.md`.
