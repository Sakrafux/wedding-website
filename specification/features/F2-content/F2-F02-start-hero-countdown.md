# `F2-F02` — Start: hero, greeting, countdown

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F01`

## Story

As a guest, I want the first screen after logging in to show our photo, our names, the date and how long it is until the wedding, so that I know I am in the right place and what it is for.

## Scope

**In:**

- `/start`, replacing the placeholder page in `web/src/routes/_guest/start.tsx`.
- `HeroSection`: photo, names, date, venue town.
- The greeting, addressed to the household by `display_name`.
- `CountdownBadge`, and what it renders on and after the wedding day.
- The primary call to action into the RSVP form.
- `web/src/lib/wedding.ts`: the wedding date as one exported constant.

**Out:**

- Navigation → `F2-F01`.
- The RSVP form itself → `F3-F01`.
- Final photo selection and crop — an open input, see below. The story ships with the real component and whatever image exists.

## Instructions

1. The page currently says "Ihr seid angemeldet" and points at this story from two places: its own doc comment, and `shellLabels.startHeading` / `startIntro` in `labels.ts`. Delete both label entries with the placeholder, and rewrite the doc comment. Leaving a `startIntro` nobody renders is how a label file grows strings that lie.
2. The wedding date is hardcoded, like all static content, but in **one** place: `web/src/lib/wedding.ts`. `labels.ts` already spells "17.07.2027" into `householdLabels.ageHint`; that string becomes a function of the constant in this story, so a date change is one edit rather than a search.
3. Greet the household by name from `me.household.display_name`, which the app already has — no new query. Free text like "Luki & Paddi" is a valid display name, so the sentence must read correctly with a name that is not "Familie …": build the copy around the name rather than gluing a salutation in front of it.
4. Hero per [05-design](../../05-design.md): `object-cover`, `aspect-[4/5]` on mobile and `aspect-[16/9]` on desktop, a warm overlay so the display text keeps its contrast against whatever the photo does, explicit `width`/`height` to prevent layout shift, `.webp` with a `.jpg` fallback. Names in `display`, and nothing else at that size on the page.
5. The hero photo is `alt=""` — it is decoration next to text that already says who and when. A description of our own engagement photo helps no screen reader user.
6. Do not lazy-load the hero. It is the first thing above the fold; `loading="lazy"` there costs a visible delay for a guest on a train.
7. The countdown shows **days**, from `wedding.ts`, and never a ticking seconds counter — explicitly rejected in [05-design](../../05-design.md). It is a badge, not a clock, and a live counter is motion for nothing.
8. **One countdown only, to the wedding.** The RSVP deadline appears on this page as a sentence with the date written out (formatted from `rsvp_deadline` in the `me` response, so moving the setting moves the sentence) next to the answer link, never as a second counter: two counters compete for the same attention and the urgent one — the deadline — loses to the bigger number. Decided 2026-08-31.
9. Three states, all of them written: days remaining, "heute" on the day itself, and hidden entirely afterwards. A countdown reading "-3 Tage" in August 2027 is the kind of thing nobody tests and everybody sees.
10. Compute the countdown against the local date at midnight, not against a timestamp difference. A guest opening the site at 23:50 must not be told there is one day less than the person next to them at 00:10.
11. One primary action, into the RSVP form, with wording that follows whether the household has answered. Secondary links belong to the nav; a start page with six equal buttons is a menu with a photo on it.
12. Respect `prefers-reduced-motion`: whatever the hero does on load, it does nothing under that query.
13. **The hero carries names and the date only** — no venue town. Decided 2026-08-31: the venues get their own page (`F2-F04`), and a start page that lists a town invites a guest to plan travel from the one line that has no address on it. **Open input, tracked in [TODO.md](../../../TODO.md):** the final photo and its crop; ship with the placeholder image checked in.
14. **Open input:** the wedding date is a working assumption (2027-07-17) until the venue is confirmed. That is exactly why it is one constant.

## Test plan

- [ ] Component: `renderApp("/start")` renders the household's display name from the `me` fixture.
- [ ] Component: the countdown renders days remaining for a date in the future.
- [ ] Component: on the wedding day it renders the "heute" wording.
- [ ] Component: after the wedding day the badge is absent.
- [ ] Component: the countdown does not change when the clock is moved within the same local day.
- [ ] Component: the RSVP call to action links into the RSVP route and reflects the answered state.
- [ ] Accessibility: exactly one `<h1>`, no skipped heading levels, hero image `alt=""`, text over the photo measured at ≥ 4.5:1.

## Done when

- [ ] A guest logging in for the first time sees the photo, their household's name, the date and the days remaining, and knows where to answer.
- [ ] `grep -rn "F2-F02" web/src` is clean.
- [ ] Checkbox ticked in `README.md`.
