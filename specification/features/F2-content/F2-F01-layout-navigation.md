# `F2-F01` — Layout, navigation, bottom bar

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F1-F03`, `F3-F01`

## Story

As a guest, I want the same visible navigation on every page, so that I can find the schedule, the venue and my answer without going back to where I came from.

## Scope

**In:**

- The guest chrome inside the `/_guest` layout: header, navigation, and the page container the content pages render into.
- A fixed bottom bar on mobile with icon + label, a horizontal top nav on desktop.
- A "Mehr" page at `/mehr` for the overflow entries.
- The persistent RSVP entry point, which changes wording once the household has answered.
- Nav entries gated by the flags from `me` (`seating_published`, `gallery_visible`).
- The German guest URLs the content stories then fill: `/start`, `/ablauf`, `/location`, `/dresscode`, `/geschenke`, `/faq`, `/kontakt`, `/mehr`, and the public `/datenschutz`.

**Out:**

- The content of any page → `F2-F02`–`F2-F07`.
- The admin shell, which has its own nav and its own density → `F1-F04`.
- The accessibility sweep across the whole app → `F11-01`. This story owns the rules for its own chrome; it does not audit anything else.

## Instructions

1. The chrome lives in the `/_guest` layout (`web/src/routes/_guest.tsx`), **not** in `__root.tsx`. The root also renders the login screen, the admin area and `/datenschutz`, and a nav bar offering "Ablauf" to somebody who has not logged in yet is both wrong and a small leak of what exists behind the door.
2. `web/src/routes/__root.tsx` carries two comments naming this story — one on the route doc comment, one beside the skip link. Both must be updated when this ships: the skip link stays where it is and keeps its reason, but it now skips past navigation that exists. A comment describing work that is finished is a stale comment, and nothing else will prompt anybody to look at it.
3. Mobile is a fixed bottom bar, per [05-design](../../05-design.md): 4–5 items, **icon plus visible German label**, never icon-only. Rejected again here, not reopened: a hamburger hides exactly the links an unconfident guest is hunting for.
4. The five bottom-bar entries are Start, Ablauf, Location, Antwort (the RSVP route `/zusagen`, owned by `F3-F01`), Mehr. Everything else lives on `/mehr`. Antwort sits in the bar rather than on the overflow page because it is the one thing we actually need the guest to do.
5. The RSVP entry is visually primary while the household has not submitted, and reads "Antwort ändern" afterwards. Both states come from data the app already has — `me` and the RSVP query from `F3` — and neither is a new field.
6. `/mehr` is a real page of large tap targets, one per entry: Dresscode, Geschenke, FAQ, Kontakt, Datenschutz, plus Sitzplan and Galerie when their flags are on. A list of links at 48px each beats a menu that overlays the page and has to be dismissed.
7. Flag-gated entries are **absent**, not disabled. An unpublished seating plan is not "coming soon", it is nothing the guest should think about yet — and the admin placeholder pattern from `F1-F04` exists for us, not for guests.
8. Desktop is the same set of entries as a horizontal top nav. One nav definition rendered twice, never two lists that drift apart.
9. The active entry gets `aria-current="page"` and a non-colour marker as well as its `primary` tint — colour is never the only signal.
10. Touch targets are at minimum 48×48px with 8px between them, and the bar respects `env(safe-area-inset-bottom)` so the last item is not under a home indicator. The page container gets bottom padding equal to the bar's height; content that ends underneath a fixed bar is content nobody scrolls to.
11. Icons come from `lucide-react`, which is already a dependency. No new package for this story.
12. Every German string goes into `labels.ts` as a `navLabels` map. The nav is the most-read text in the product; it does not get to be inline strings.
13. Content pages render inside one shared container: `max-w-2xl`, centred, 16px gutter on mobile and 24px on desktop, 48/64px between sections. The content stories then write prose, not layout.
14. `shellLabels.logout` moves out of the start page and into the chrome — on `/mehr` rather than in the bar. It is used once a year, and a fifth bar item that logs you out next to the one that shows the schedule is a mis-tap waiting to happen.

## Test plan

- [ ] Component: `renderApp("/start")` shows the bottom bar with all five entries, each with a visible label.
- [ ] Component: tapping an entry navigates, and the target entry carries `aria-current="page"`.
- [ ] Component: the RSVP entry reads the "noch nicht geantwortet" wording before submission and "Antwort ändern" after it.
- [ ] Component: with `seating_published: false` and `gallery_visible: false` the Sitzplan and Galerie entries are absent from `/mehr` — asserted by their absence, so a later change that renders them disabled fails here.
- [ ] Component: `/mehr` links to Datenschutz, and that link still resolves when logged out.
- [ ] Component: the login screen and the admin shell render **no** guest nav.
- [ ] Accessibility: the nav is a `<nav>` with an accessible name, the skip link still lands on `#main`, and every entry is reachable and operable by keyboard with a visible focus ring.

## Done when

- [ ] Every guest page is reachable from every other guest page without the browser's back button.
- [ ] The `F2-F01` comments in `__root.tsx` are gone or rewritten; `grep -rn "F2-F01" web/src` is clean.
- [ ] Checkbox ticked in `README.md`.
