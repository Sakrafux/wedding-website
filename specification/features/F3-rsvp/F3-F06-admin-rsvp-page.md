# `F3-F06` — Admin RSVP page, reusing the guest form

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-B06`, `F3-F05`, `F5-F02`

## Story

As an admin, I want to fill in a household's answer on the same form they would have used, so that a phone call takes two minutes and no field goes missing because the admin form was built separately.

## Scope

**In:**

- `/admin/haushalte/{id}/rsvp`.
- The **same** `RsvpForm` component, given the admin query and the admin mutation as props.
- The admin-only framing around it: whose answer this is, when they last changed it, that the deadline does not stop us.
- Replacing the read-only RSVP summary on the household detail page with a link here.

**Out:**

- A second form. If this story renders anything the guest form does not, that thing belongs in the shared component or nowhere.
- Note triage (seen/unseen) → `F6-F02`.

## Instructions

1. Render `RsvpForm` with `data` from `GET /api/admin/households/{id}/rsvp` and `onSave` bound to the `PUT`. No other difference. This is what `F3-F01`'s props-not-fetching rule was for.
2. The page header names the household and states plainly that you are answering **for** them. An admin who forgets which of two tabs is which will otherwise write Familie Müller's answer into Familie Schmidt.
3. When the deadline has passed, show it as information, not as a lock: the guest form would be read-only here, and this page must not be. `editable: false` from the API therefore does **not** switch this page into the `F3-F05` view — pass the override explicitly, so the shared component's default stays the safe one.
4. After saving, the same summary renders. It is the recap we read back over the phone, which is the reason not to replace it with a toast here either.
5. `F5-F02` currently shows a read-only RSVP summary with a comment saying this story replaces it. **Resolve that forward reference**: replace the placeholder with a link, and delete the comment at `web/src/routes/admin/_shell/haushalte/$householdId.tsx` and the note in `labels.ts` that names `F3-F06`. Run `grep -rn "F3-F06"` before ticking the box.
6. Admin density: drop the guest page's decorative spacing, keep every control and every touch target. It is the same component; do not fork it to make it denser.
7. A 404 for an unknown household lands on the admin household list with a message, not on the guest login.

## Test plan

- [ ] Component: the page renders the same form component as the guest route — assert by shared component, not by duplicated markup expectations.
- [ ] Component: with `editable: false`, the admin page still renders inputs, and the guest page does not.
- [ ] Component: saving posts to the admin endpoint with the household id from the path.
- [ ] Component: the summary renders after a successful save.
- [ ] Component: the household detail page links here and no longer renders a read-only RSVP block of its own.
- [ ] Component: a 401 lands on the admin login, not the guest one.

## Done when

- [ ] A phoned-in answer can be entered in one pass, after the deadline, and `grep -rn "F3-F06"` returns nothing outside the specification.
- [ ] Checkbox ticked in `README.md`.
