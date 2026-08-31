# `F2-F04` — Location, Anreise & Übernachtung

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F01`

## Story

As a guest, I want the addresses of both venues with how to get there, park, and stay the night, so that I can plan the journey myself.

## Scope

**In:**

- `/location`, one page with three sections: the two venues, Anreise, Übernachtung.
- An address block per venue: name, street, postcode and town, plus an external map link.
- Parking, public transport, and the church → reception trip.
- A short list of nearby accommodation.

**Out:**

- Matching drivers to riders. The app never does this — the RSVP form collects seats needed and offered (`F3-F03`) and we arrange the rest offline.
- The schedule → `F2-F03`.

## Instructions

1. One page, not three. A guest asking "where is it" wants the address, the parking and the train in the same scroll; splitting them means answering the same question on three pages and hoping they find the right one. Sections get `id` anchors so the nav or an FAQ answer can link straight at one.
2. Both venues get the same address block component. The church and the reception are two places with the same set of facts, and two hand-written blocks are two places for a postcode to be wrong.
3. Map links are **plain external links** to a map service, opening in a new tab. No embedded map: an iframe is a third party in the critical path, it needs a `frame-src` hole in the CSP from `E0-07`, and it hands every guest's IP to a mapping company. That trade is refused for a link that does the same job.
4. Mark external links as such, in text as well as by icon, and use `rel="noopener noreferrer"`. Guests in this audience should not be surprised by a new tab.
5. Addresses are marked up so a phone offers to open them: a real address block with the parts on separate lines, and the map link doing the actual work. Do not chase microformats — the link is the affordance that works.
6. Say plainly how the trip from the church to the reception is meant to work, and that the RSVP form is where seats are asked for. Otherwise the transport question in the form arrives without context, and a guest who has already answered it here has nowhere to put the answer.
7. Accommodation is a short list of names with a link each, and one sentence saying we have not blocked rooms unless we have. Anything more is a travel agency.
8. Prose is capped at `max-w-prose`; long addresses do not wrap mid-postcode.
9. **Open input, tracked in [TODO.md](../../../TODO.md):** venue names, addresses, parking and travel notes, and the hotel list. [07-roadmap](../../07-roadmap.md) is explicit that a thin Ablauf page is acceptable at launch and an **empty Location page is not** — the venue is the first thing a guest looks for. This page therefore blocks send-out; it is not one to ship as a placeholder.
10. The content in this page ships inside the JavaScript bundle, and the bundle is served to anyone who loads the site, logged in or not — the session gate is on `/api`, not on the SPA. That is fine for a venue address on an invitation, but note it here so the next person adding something to a content page knows the audience is "anyone with the URL", not "our guests". See `F2-F05`, where it actually bites.

## Test plan

- [ ] Component: `renderApp("/location")` renders both venue blocks with their addresses.
- [ ] Component: the map links point at the configured URLs, carry `rel="noopener noreferrer"`, and are marked as external.
- [ ] Component: the section anchors exist and a link to `#anreise` scrolls to that section.
- [ ] Accessibility: sections are `<section>` with headings in order, no skipped levels, and every link's accessible name says where it goes rather than "hier".

## Done when

- [ ] A guest can find both venues, park, arrive by train and book a bed without calling us.
- [ ] Checkbox ticked in `README.md`.
