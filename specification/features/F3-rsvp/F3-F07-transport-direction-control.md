# `F3-F07` — Transport as one question with a direction

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-B07`, `F3-F03`

## Story

As a guest, I want to say in one step whether we need seats or can offer them, so that I am not answering two questions that look like the same question asked twice.

## Scope

**In:**

- A three-way choice — nothing, we need seats, we can offer seats — followed by one count.
- The count steppers from `F3-F03`, now one at a time.

**Out:**

- The server-side rule → `F3-B07`. This story makes the conflict unreachable; it does not replace the refusal.
- `has_stroller` and the note, which stay exactly as `F3-F03` built them.

## Instructions

1. The direction is **derived from the two counts**, never stored as a third piece of state: `needed > 0` is "we need", `offered > 0` is "we offer", both zero is "nothing". A separate mode field would be a fourth thing to keep in step with a body that has two numbers in it.
2. Choosing a direction zeroes the other count and starts the chosen one at 1. Starting at 0 would leave a household who picked "wir brauchen Plätze" looking at a zero and a save button, which is a form telling somebody their answer does not count.
3. `RadioCardGroup`, like every other guest-facing choice: three cards, one stepper below the chosen one. The stepper's minimum is 1 while a direction is chosen — going back to nothing is what the third card is for, and it says so in words.
4. The section stays gated on somebody attending `both`, and `transportDropped` still explains a vanished answer (`F3-F03`).
5. Help copy moves with the question: one popover on the direction saying this is the church → reception trip and that we are deciding whether a shuttle is worth organising, and one on the count. Two steppers that each needed their own *why* are what this control removes.

## Test plan

- [ ] Component: a household with `needed = 2` renders with "wir brauchen Plätze" selected and the count at 2.
- [ ] Component: switching to "wir bieten Plätze" zeroes `needed` and submits `offered = 1`.
- [ ] Component: "wir brauchen nichts" submits both as zero.
- [ ] Component: the stepper does not go below 1 while a direction is chosen.
- [ ] Accessibility: the radio group is labelled, and the stepper's buttons name the field they change.

## Done when

- [ ] A household states its transport situation once, and cannot express a contradiction.
- [ ] Checkbox ticked in `README.md`.
