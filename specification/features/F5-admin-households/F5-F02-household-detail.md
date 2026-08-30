# `F5-F02` — Household detail and member editing

**Epic:** F5 — Admin: households & guests · **Layer:** frontend · **Depends on:** `F5-B02`, `F5-B03`, `F5-F01`

## Story

As an admin, I want one page per household where I can fix the name, add and correct the people in it, and reissue the code, so that entering eighty guests is a single flow rather than four screens.

## Scope

**In:**

- `/admin/haushalte/{id}`.
- Household fields: name, private note, transport seats, stroller.
- The member list, with add, edit and remove.
- The code, with copy-to-clipboard and a guarded regenerate.
- Delete the household, behind a confirmation that says what is lost.

**Out:**

- RSVP answers, which are read-only here and empty until `F3` — see the open question in `TODO.md` about entering them on a household's behalf.
- Seating → `F7`.

## Instructions

1. Consume the `F5-B01`, `F5-B02` and `F5-B03` contracts exactly. Invent no fields.
2. Entering the guest list is the bulk task this page exists for, so the add-member form stays open after a save and returns focus to the first name field. Eighty guests entered four at a time is the actual workload; a form that closes after each one triples the clicking.
3. The `age` field appears only when `kind` is `child`, and switching to `adult` clears it visibly. The server enforces the pairing (`F5-B02`); the form should not let the admin build an invalid combination and then read about it.
4. Member edits save per row on blur or on an explicit save, not through a single page-wide submit. A page-wide submit over twenty members means one validation error discards the whole screen.
5. Show `origin` for `guest_added` members with the label from `labels.ts`; `seeded` shows nothing. A marker on every ordinary guest drowns out the one that matters.
6. The private note is labelled unambiguously as private — the admin should never wonder whether a household can read it. It cannot: `dto.HouseholdSummary` omits it, and `F1-B04` asserts that.
7. The code is displayed in printed form with a copy button. Copy copies the **printed** form, since that is what gets pasted into a document that goes to the printer.
8. Regenerating is behind a confirmation naming the consequence in plain German: the old code stops working, any printed card carrying it is dead, and logged-in devices are signed out. Show `revoked_sessions` from the response afterwards, so the admin sees what happened rather than inferring it.
9. Deleting the household is behind a confirmation that states the member count being deleted with it, and requires more than one click. Its audit trail survives, and the dialog says so — the fear when deleting is losing the record, not losing the row.
10. Removing a member says "entfernen", not "löschen": it is a soft delete, the person stays in the record, and the German should not promise otherwise.
11. Optimistic updates are not worth it here. This is an admin on a laptop on a good connection, and a member list that reorders itself and then snaps back is worse than a spinner.

## Test plan

- [ ] Component: household fields render and save; a validation error appears next to its own field, in German.
- [ ] Component: adding a member keeps the form open and clears it for the next one.
- [ ] Component: the `age` field appears for `child` and disappears for `adult`.
- [ ] Component: `guest_added` members are marked; `seeded` ones are not.
- [ ] Component: regenerate is confirmed first, and the new code is shown afterwards.
- [ ] Component: delete is confirmed first and names the member count.
- [ ] Component: removing a member drops it from the list without a reload.
- [ ] Accessibility: every field has a real label; the confirmation dialogs trap focus and are dismissible by keyboard.

## Done when

- [ ] A household and its four members can be entered from scratch in one pass, without leaving the page.
- [ ] Checkbox ticked in `README.md`.
