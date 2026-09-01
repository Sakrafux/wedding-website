# `F3-F09` — The stored answer is what `/zusagen` shows once answered

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-F04`

## Story

As a guest who has already answered, I want `/zusagen` to show me what we told you, so that I am not re-reading a filled-in form to work out whether it was saved.

## Scope

**In:**

- `/zusagen` renders the summary when `rsvp_submitted_at` is set, with "Antwort ändern" leading to the form.
- Saving returns to that same summary, with a confirmation above it.
- The admin RSVP page stays form-first.

**Out:**

- The read-only view after the deadline, which is a different state with different copy → `F3-F05`.

## Instructions

1. One view state with two entry points, replacing the post-save-only summary. Reachable only by saving, the summary was a screen that appeared once and could never be found again — which is also why a household could not tell whether their last edit had been stored.
2. The summary renders from the query cache, which the save writes into (`useSaveRSVP`), so what is shown after saving is the stored, normalized answer and not the draft.
3. The confirmation heading stays the focus target after a save, as `F3-F04` built it. Arriving at the summary by navigation moves no focus — nothing happened that a screen-reader user needs told.
4. The admin form is form-first, deliberately: we open that page mid-phone-call to change something, and a summary would be one tap between us and the field. The prop that decides it defaults to form-first, so the guest route is the one that opts in.
5. The nav label ("Antwort ändern" once answered, `F2-F01`) already matches this view — a guest who arrives from the bar and one who arrives from the start page see the same page.

## Test plan

- [ ] Component: an answered household lands on the summary; "Antwort ändern" reveals the form.
- [ ] Component: an unanswered household lands on the form.
- [ ] Component: saving from the form shows the confirmation and the summary of the *saved* answer.
- [ ] Component: the admin page renders the form even for an answered household.

## Done when

- [ ] A household can see what we have on record without pressing save to find out.
- [ ] Checkbox ticked in `README.md`.
