# `F11-01` — Accessibility pass: keyboard, focus, contrast, 200% zoom

**Epic:** F11 — Cross-cutting quality · **Layer:** frontend · **Depends on:** `F2-F07`, `F3-F05`, `F4-F02`, `F5-F03`

## Story

As a guest who is 75, uses a phone at 150% system font scale and taps with one thumb, I want every screen to be operable as it stands, so that answering the invitation does not require phoning us.

## Scope

**In:**

- An audit of every screen that exists at M3 — login, confirmation, the F2 content pages, the RSVP form, the add-member sheet, and the admin screens — against the accessibility rules in [05-design](../../05-design.md).
- The fixes the audit turns up.
- The checklist itself, as a table in this file, so `F6`–`F10` are audited as they are built instead of accumulating a second pass.

**Out:**

- Screens that do not exist yet. `F6`, `F7`, `F8`, `F9` and `F10` run this story's checklist inside their own frontend stories; a second sweeping audit at the end is how the first one's fixes get re-broken unnoticed.
- The seating SVG's text fallback → `F7-F04`, which owns it because the fallback is a feature and not a remediation.
- German wording and error copy → `F11-02`.
- Real-hardware verification → `F11-03`. This story is desk work in a desktop browser; that one is the phone in a relative's hand.

## Instructions

1. This is an **audit-and-fix** story, not a "make it accessible" story. The rules were built in from `E0-08` onward and every frontend story carries accessibility bullets in its own test plan. What a single pass catches is *drift* — the control added in a hurry, the heading level skipped, the marker that ended up colour-only. Treat a finding as a defect in the story that introduced it, not as work this story was always going to do.
2. **No automated accessibility checker.** `axe-core` in jsdom was considered and rejected: it cannot see touch-target size, reflow at 200% zoom, whether a focus ring is visible, whether a colour-coded marker also carries text, or the register of the German copy — which is most of what this audience actually needs — and a green automated run reads as a finished audit. The screens listed above are walked by hand instead, against a checklist that says what was checked.
3. The checklist lives **in this story**: one table, a row per check, a column per screen. A separate `specification/08-accessibility.md` was considered and rejected — the rules themselves are already in [05-design](../../05-design.md), so a second document would restate them and then disagree with them, and a table of tick marks is progress, which `CLAUDE.md` says belongs in `features/README.md` and nowhere else. `F6`–`F10` cite `F11-01` and add their screens as columns; citing a finished story is what every other epic here already does.
4. Screen readers are **out of scope**. Nobody in this guest list uses one, and a VoiceOver/TalkBack pass done badly by someone who does not use the tool daily produces false confidence rather than findings. Semantic HTML, real labels and keyboard operability are still required — they are what make the page work at 200% zoom and under a thumb, which is the real risk here — and they happen to be most of what a screen reader would want anyway.
5. Manual pass, per screen, in a real browser at a 360px viewport:
   - Keyboard only, no mouse: reach every control, in an order that matches the visual one, with a visible 2px `primary` focus ring at 2px offset on each. Nothing focusable that is not operable, nothing operable that is not focusable.
   - Skip-to-content link is the first focusable element and actually moves focus.
   - One `<h1>`, no skipped heading levels, and a document title per route — a browser tab and a screen-reader announcement both read it, and TanStack Router will happily leave every route sharing one title.
   - Browser zoom 200% and OS font scale 1.5×: no horizontal scrolling, no clipped or overlapping text, the bottom bar from `F2-F01` still shows its labels.
   - Touch targets measured, not eyeballed: 48×48px minimum with 8px between neighbours. The bottom bar, the RSVP radio cards and the member-card controls are where this fails first.
   - `prefers-reduced-motion: reduce` set: nothing moves.
6. Contrast is **measured on rendered pixels**, not taken from [05-design](../../05-design.md). The table there is correct about the tokens; what this checks is that no component overrode one — `ink-muted` on `surface-sunken` rather than on `paper`, `accent` used as text where `accent-strong` was meant, placeholder grey, disabled states.
7. Walk the colour-only list deliberately, because it is the rule most easily lost in a later edit: "nie angemeldet" (`F5-F01`), the `guest_added` marker (`F5-F02`), validation errors, the RSVP saved confirmation, and the soft-cap hint (`F4-F01`). Each needs an icon or a word as well as its colour.
8. Semantic HTML first. If a finding's fix is an ARIA attribute, check whether the real problem is a `<div>` that should have been a `<button>` — `aria-*` patching a wrong element is how a screen reader ends up announcing something the keyboard cannot reach.
9. Fix what the audit finds. Anything deliberately not fixed goes into root `TODO.md` with the reason and the screen, because an unrecorded known defect is indistinguishable from one nobody noticed.

## Test plan

- [ ] Component: the skip link is the first focusable element and moves focus to the main landmark.
- [ ] Component: every route renders exactly one `<h1>` and sets a distinct document title.
- [ ] Negative: a control deliberately given a 32px hit area is caught by the manual measurement — a check on the *checklist*, confirming it is doing work rather than being ticked.
- [ ] Manual: the per-screen checklist is filled in for every M3 screen, keyboard, zoom, font scale, targets and contrast.

## Done when

- [ ] Every M3 screen has a completed checklist, and every finding is either fixed or recorded in `TODO.md` with a reason.
- [ ] The checklist table in this file is filled in for every M3 screen, and is what `F6`–`F10` extend.
- [ ] Checkbox ticked in `README.md`.
