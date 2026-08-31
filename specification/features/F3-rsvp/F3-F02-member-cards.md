# `F3-F02` — Member cards with scope-gated catering fields

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-F01`

## Story

As a guest, I want to be asked about food only for the people who are actually coming to the party, so that answering for a grandmother who only comes to the church is three taps and not twelve.

## Scope

**In:**

- `RsvpMemberCard`: name, attendance scope, and — revealed only for `party_only` and `both` — meal choice, portion, midnight snack.
- `seating_need` and `dietary_note`, shown for anybody who is attending at all.
- A child's `age`, asked as age **at the wedding date**.

**Out:**

- Household-level fields → `F3-F03`. Submit → `F3-F04`.
- Adding or removing a member → `F4-F01`, `F4-F02`.

## Instructions

1. Consume the `F3-B02` member shape exactly. Invent no fields.
2. The catering fields are **revealed**, not disabled, when the scope covers the party. Disabled controls fail contrast and read as broken (`05-design`); an absent question reads as a question that was not asked, which is exactly what it is.
3. Reveal is animated only as far as `prefers-reduced-motion` allows, and the revealed block is announced — the fields appear below the control that caused them, never above, so the page does not jump under a thumb.
4. Scope `no` collapses everything except the name and the scope control. Nothing else is asked of somebody who is not coming; a declining household should be finished in four taps.
5. `seating_need` and `dietary_note` are shown for `church_only` too. A wheelchair space is needed in the pew, and this is the one place the scope gate deliberately does not apply — say so in a comment, because it looks like an inconsistency.
6. Options render from `labels.ts` — `attendingLabels`, `mealChoiceLabels`, `portionLabels`, `portionHelpLabels`, `seatingNeedLabels`. The help text under `portion: none` is the reason that map exists.
7. Meal choice and portion are radio cards, not a `<select>`. Three options each, and a native select on Android hides them behind a modal for no benefit.
8. Midnight snack is a checkbox with a one-line explanation of what it is — the term means nothing to a guest who has not been to a wedding lately.
9. `age` appears only for `kind = 'child'`, labelled as age **on the wedding day**, with the date in the label. The wording is the whole reason the value does not drift; a bare "Alter" gets answered as of today.
10. `kind` is not editable here. A household that needs to change it phones us and we fix it in `F5-F02` — see `F3-B01`.
11. Each card is a `Card` with the member's name as its heading, and every control has a real `<label>` that names the person: "Was isst Anna?" beats a bare "Essen" repeated eight times down the page, for anybody navigating by label.
12. Field errors render under their own control with `aria-invalid`, keyed from `members.<id>.<field>`.
13. **Every field carries a help popover** — the `?` button beside its label, per `05-design`'s form behaviour rules. On this card that is the scope control, meal choice, portion, midnight snack, seating need, dietary note and a child's age: seven fields, none of which is self-explanatory to somebody who has not planned a wedding. The German sentences live in `labels.ts` beside the enum labels they explain, and `portion: none` keeps its inline hint as well, because that one disambiguates two options rather than describing a field.
14. **Nothing is marked as an error before the guest has tried to save.** A form that opens red at a household who has not answered yet reads as broken, and it is the opposite of the register the copy is written in. Validation is on blur and on submit (`05-design`), and an unanswered member carries the neutral "Noch keine Antwort" marker from `labels.ts` — a statement of fact, in `ink-muted`, not `danger`. Only after a submit attempt does an unanswered card become an error (`F3-F04`). Decided 2026-08-31.

## Test plan

- [ ] Component: choosing `both` reveals meal, portion and snack; choosing `church_only` hides them; choosing `no` hides everything but the scope.
- [ ] Component: switching from `both` to `church_only` and back does not resurrect a stale meal choice in the submitted body — the client sends what is on screen, and the server normalizes regardless.
- [ ] Component: `seating_need` and `dietary_note` remain visible for `church_only`.
- [ ] Component: the age field appears only for children and names the wedding date.
- [ ] Component: a field error from the API renders under the right control of the right member.
- [ ] Component: a freshly loaded form with no answers shows no error styling anywhere — the unanswered marker is neutral until a submit is attempted.
- [ ] Component: every field has a help button whose accessible name says which field it explains, and whose popover closes on Escape with focus returning to the button.
- [ ] Accessibility: every control has a label naming its member; the revealed block does not steal focus.

## Done when

- [ ] A household with a party guest, a church-only guest and a child can be answered correctly, and nobody is asked a question that does not apply to them.
- [ ] Checkbox ticked in `README.md`.
