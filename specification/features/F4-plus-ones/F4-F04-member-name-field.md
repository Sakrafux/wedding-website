# `F4-F04` — Name field on the member card

**Epic:** F4 — Plus-one · **Layer:** frontend · **Depends on:** `F4-B04`, `F3-F02`

## Story

As a guest, I want the name on each card to be a field I can edit, so that correcting "Erika Huber" to "Oma Erika" takes a keystroke rather than a phone call.

## Scope

**In:**

- A `name` input as the first field of every member card, with the `?` popover the other fields have.
- `name` in `MemberDraft`, seeded from the response and serialised into the save.
- Every label on the card addressing the person by the **typed** name.
- "Eingeladen als …" on the admin household detail page, only when the household has renamed somebody.

**Out:**

- Editing the household's own name or addressee. Admin-only (`F5-F01`).
- A separate save or a pencil affordance. The card's name saves with the rest of the form.
- Any undo. The value we invited them under is on the admin page, and that is the whole remedy.

## Instructions

1. The field is first on the card, above the scope control: it names who the rest of the card is about, and a label reading "Wozu kommt …?" above the thing that decides the name reads backwards.
2. `MemberDraft.name` is a plain string, seeded in `draftFrom` and sent by `toRequest`. A member the draft never learned about falls back to the response's name, so a save cannot blank somebody out.
3. Labels use the typed name, falling back to the stored one while the field is empty — a heading reading "Wozu kommt ?" mid-edit is worse than a momentarily stale name. Trimming for display only; the server owns the stored form.
4. `maxLength={160}`, mirroring the server's bound, so the limit is reached rather than reported.
5. German copy in `rsvpLabels` (`memberNameLabel`, `memberNameHelp`) and `householdLabels.seededName`, never inline.
6. The admin detail line reads **"Eingeladen als <Name>"**, not "früher": the question it answers is which name is on the card we posted. Rendered only when `seeded_name` is non-empty and differs from `name`, so it is silent for everybody nobody renamed.

## Test plan

- [x] Component: typing a new name re-labels the card and is what the save request carries.
- [x] Component: the admin detail page shows "Eingeladen als …" for a renamed member and nothing for an unrenamed one.
- [x] The existing "submits exactly the form state" test carries `name`, so the request shape stays asserted whole.

## Done when

- [x] A household can correct a name on the RSVP form, and the admin page still shows who we invited.
- [x] `pnpm test`, `pnpm lint` and `tsc -b` are green.
- [x] Checkbox ticked in `README.md`.
