# `F3-F01` — Form shell and household scope selector

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-B02`, `F1-F03`

## Story

As a guest, I want one page that asks who is coming, starting with a single question for the whole household, so that a family of four answers in one tap and only edits the exceptions.

## Scope

**In:**

- The guest route `/zusagen`, under the `_guest` layout.
- `RsvpForm`: the component that owns the form state for the household and every member.
- `ScopeSelector` — the household-level "Wir kommen zu:" bulk setter.
- Loading, error and empty states through the existing `RouteStates` components.

**Out:**

- The per-member fields → `F3-F02`. Transport and note → `F3-F03`. Save, summary → `F3-F04`. Read-only → `F3-F05`.
- Adding a member → `F4-F01`.

## Instructions

1. **`RsvpForm` takes its data and its save mutation as props and never fetches for itself.** The route component does the fetching. This is what lets `F3-F06` render the identical component against the admin endpoint, and it is the constraint the whole epic is arranged around — see `TODO.md`, decided 2026-08-31. A `useQuery` inside this component is the one change that breaks the admin page.
2. Consume the `F3-B02` contract exactly. Invent no fields.
3. Route path `/zusagen`, not `/rsvp`: "RSVP" is jargon the copy rules forbid, and the URL is read aloud on the phone as often as it is clicked.
4. One long form with clear sections, not a wizard. Rejected in `05-design` and re-stated here because a member-per-step wizard is the obvious thing to reach for with eight guests on one page: it hides the length and makes fixing one answer a journey.
5. `ScopeSelector` is large radio cards, minimum 48×48px, with the four options labelled from `attendingLabels`. It is a **bulk setter**, not a stored field: choosing one writes that scope into every member card below, and the cards remain individually editable afterwards.
6. Once any member differs from the household selection, the selector shows no option as chosen and says so in plain German ("Ihr habt unterschiedliche Antworten"). A selector that stays lit while the cards disagree is a lie about what will be saved.
7. Choosing a household scope after the cards were edited **overwrites** them all. Guard it with a confirmation naming the count that will change, and only when something actually would change. Silent bulk overwrite is how a household loses Oma's church-only answer.
8. The form is one `<form>` with a single submit. Nothing autosaves (`05-design`); the moment of completion is the whole point.
9. State lives in one object keyed by member id, mirroring the request shape of `F3-B03`, so submitting is a serialisation and not a gathering exercise. Field errors from the API arrive keyed `members.<id>.<field>` — the same key — so mapping an error to a card is a lookup.
10. Heading structure: one `<h1>` ("Sagt uns Bescheid"), `<h2>` per section. No skipped levels; the page is long and a screen-reader user navigates it by heading.
11. The route is the primary call to action in the navigation until `rsvp_submitted_at` is set, then becomes "Antwort ändern" (`05-design`). `F2-F01` owns the bar; this story provides the flag it reads, from the `F3-B02` response.
12. Every field on this page and the ones below it carries a help popover behind a `?` button beside its label (`05-design`). The shared component that renders it lands here, with the rest of the form shell, rather than being invented three times in `F3-F02`, `F3-F03` and `F4-F01`.
13. All copy goes in `web/src/lib/labels.ts`, help sentences included. No German string inline.

## Test plan

- [ ] Component: the form renders every member from the API response, in order.
- [ ] Component: choosing a household scope sets every member card to it.
- [ ] Component: editing one member clears the household selection rather than leaving it lit.
- [ ] Component: re-choosing a household scope after edits asks for confirmation and names the count.
- [ ] Component: a household with one member renders without the bulk selector looking absurd — either it is hidden or it is the only control; decide and assert it.
- [ ] Component: an API failure on load renders the shared error state, not a blank form.
- [ ] Accessibility: the scope options are a real radio group with a group label, reachable and operable by keyboard.

## Done when

- [ ] A household of four can set "Kirche und Feier" for everyone in one tap, and the form is otherwise ready for the member fields.
- [ ] Checkbox ticked in `README.md`.
