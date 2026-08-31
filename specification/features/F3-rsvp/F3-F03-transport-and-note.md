# `F3-F03` — Transport seats and household note

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-F02`

## Story

As a guest, I want to say how many seats we need or can offer for the drive to the reception, and to write down anything the form did not ask, so that we are not left phoning about the one thing that matters to us.

## Scope

**In:**

- `TransportFields`: seats needed and seats offered, with plain-language help.
- `has_stroller`.
- The free-text household note, labelled so it is obvious that a human reads it.

**Out:**

- The capacity gap the counts feed → `F6`. The guest never sees a total.
- Matching riders to drivers — not a feature, here or anywhere (`02-features`).

## Instructions

1. Consume the `F3-B02` household shape exactly. Invent no fields.
2. The transport section is shown only when at least one member is attending `both`. The trip is church → reception; nobody else makes it, and the server zeroes the counts in that case anyway (`F3-B01`). Hiding it keeps two questions off the screen of every household that is only coming to one half.
3. If the section is hidden after having had values — somebody changed their scope — say so rather than silently discarding: a one-line note that the transport answers no longer apply. A number that vanishes without explanation is the kind of thing that gets phoned in about.
4. Steppers with `−` and `+` buttons at 48×48px, not a number input. This audience mis-taps far more often than it mistypes, and a spinner on mobile Safari is a two-pixel target.
5. Both seat fields carry a help popover (`05-design`), and they are the clearest case for one: without the *why*, "gebraucht" and "angeboten" read as the same question asked twice. The popovers say that this is the church → reception trip, and that the offered seats tell us whether to organise a shuttle. `has_stroller` and the note get one each too.
6. `has_stroller` is a single checkbox in the same section, phrased as a question about the pram rather than about a person — it is a household fact, which is why it is not on a member card.
7. The note is a `Textarea` with a heading that invites use: this is the deliberate catch-all, and a field labelled "Anmerkungen" gets left empty. Give an example of the kind of thing that belongs in it, taken from `02-features` ("wir kommen erst nach der Zeremonie", "Oma braucht einen Platz nah am Ausgang").
8. Say plainly that we read it. No promise of a reply, no character counter until the 2000-character cap is close.
9. The note is one field for the household, not per member, and sits at the end of the form — after the members, before the save button.

## Test plan

- [ ] Component: the transport section is hidden when nobody attends `both`, and appears when somebody does.
- [ ] Component: hiding it after values were entered shows the explanation.
- [ ] Component: the steppers do not go below zero and stop at the cap.
- [ ] Component: the note round-trips through the form state and is submitted as `rsvp_note`.
- [ ] Component: each field's help popover opens from its own button and names its field in the accessible name.
- [ ] Accessibility: the steppers are buttons with accessible names that include the field, not bare `+` and `−`; the value is announced on change.

## Done when

- [ ] A household can state its transport situation and write us a sentence, and no household is asked about a drive it is not making.
- [ ] Checkbox ticked in `README.md`.
