# `F3-F05` — Read-only rendering after the deadline

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-F04`

## Story

As a guest, I want to still see what we answered after the deadline, with a clear explanation that it can no longer be changed, so that I am not left wondering whether the form is broken.

## Scope

**In:**

- The read-only rendering of the whole RSVP page when `editable` is false.
- The explanatory `Alert`, including what to do if something has changed.
- The navigation label after the deadline.

**Out:**

- Server enforcement → `F3-B04`. This story renders a state; it does not create one.
- What the rest of the site says between the deadline and published seating — still open in `TODO.md`, and not this story's to invent.

## Instructions

1. Render the saved answers as **text on `surface-sunken`**, not as disabled inputs. Disabled text fails contrast and reads as broken (`05-design`); the same rule that governs the summary in `F3-F04`, and in fact the same component — `RsvpSummary` renders both.
2. The `Alert` above it names the state and the way out: the deadline has passed, this is what we have, if something has changed ruf uns an. The phone number comes from `contactPhoneNumber` in `labels.ts`, which is **still a placeholder** (`TODO.md`) — this is the second screen that strands a guest if the real number does not land before send-out.
3. A household that never answered sees a different sentence: there is nothing to show, and the useful thing to say is to call us, not to explain a form they cannot use.
4. `editable` comes from the API (`F3-B02`), never from comparing the deadline against the browser clock. A phone with a wrong date would otherwise show a live form that every save rejects.
5. The navigation entry becomes plain rather than primary, labelled "Eure Antwort" — the persistent call to action exists to chase an answer, and after the deadline there is nothing to chase.
6. Do not hide the page. A guest who wants to check whether they said yes to the party must be able to.

## Test plan

- [ ] Component: with `editable: false`, no input is rendered anywhere on the page and the alert is present.
- [ ] Component: with `editable: false` and no answer, the "nothing recorded" variant renders with the phone number.
- [ ] Component: the read-only view uses the same summary component as the post-save state.
- [ ] Component: `editable: true` renders the form, regardless of what the browser clock says.
- [ ] Accessibility: the alert is announced, and the read-only text meets contrast on `surface-sunken`.

## Done when

- [ ] After the deadline the page explains itself without a guest having to try a save to find out.
- [ ] Checkbox ticked in `README.md`.
