# `F11-05` — Form controls look editable, and look alike

**Epic:** F11 — Cross-cutting quality · **Layer:** frontend · **Depends on:** `E0-08`

## Story

As a guest, I want a text field to look like something I can type in, so that I do not skip a question because the box looked switched off.

## Scope

**In:**

- `Input`, `Textarea` and `NativeSelect`: the surface colour, one radius, one height, one type size.
- Removing shadcn's `md:text-sm`, which is below the type floor in `05-design`.

**Out:**

- The disabled *state* itself, which stays as it is — it is used deliberately in a few places and is correct there.
- Any change to `RadioCardGroup`, `Checkbox` or `Stepper`, which were built to the design tokens already.

## Instructions

1. The three controls came in from shadcn with `bg-transparent`, which on `--color-paper` renders a cream field on a cream page — the look of a disabled input. They get `bg-surface`: white, which is what makes an editable field read as a hole in the page rather than a patch of it.
2. One radius for all three (`rounded-lg`, the token's input radius, per `05-design`) and one height. They currently disagree with each other, which is the second half of the complaint: two controls in the same form looking like two kinds of control.
3. `md:text-sm` goes. `05-design` sets 16px as the floor for this audience and states why; a desktop breakpoint that shrinks input text below the body size contradicts it, and it arrived as a vendor default rather than a decision.
4. Minimum height is the 48px touch target, not shadcn's 36px. The admin screens use the same controls and lose nothing by being comfortable.
5. No visual change to focus rings or the invalid state: both are already tokenised and both are load-bearing for accessibility.

## Test plan

- [ ] Manual, against `05-design`: a text input, a select and a textarea side by side share their radius, height and type size.
- [ ] Manual: an enabled and a disabled input are told apart at a glance.
- [ ] Component tests keep passing — this is a class-list change, and nothing queries on classes.

## Done when

- [ ] No editable field on any screen reads as disabled.
- [ ] Checkbox ticked in `README.md`.
