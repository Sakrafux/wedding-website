# `F11-03` — Mobile device QA on real hardware

**Epic:** F11 — Cross-cutting quality · **Layer:** ops · **Depends on:** `F11-01`, `F11-02`, `F2-F07`, `F3-F05`, `F4-F02`

## Story

As the person who will get the phone calls, I want the whole guest journey walked on the actual phones our guests own, over mobile data and through the real proxy, so that the failures are found by us in a living room instead of by a 70-year-old with an invitation in their hand.

## Scope

**In:**

- A scripted end-to-end walkthrough of the guest journey — card to submitted RSVP to changed RSVP — run on real devices.
- A device matrix covering the browsers our guest list actually uses, including at least one genuinely old phone borrowed from the family.
- Recorded evidence per device, and a finding list.

**Out:**

- Fixes. Findings are filed against the story that owns the screen; this story owns *finding* them, and a QA story that also does the fixing hides how much was wrong.
- The pre-print login verification → `E-OPS-03`. See the gate note below.
- The admin screens. They are used by one person on one laptop, who will notice.

## The gate note, because it is the thing that gets confused

**`E-OPS-03` gates the print run**, not this story. That gate is narrow on purpose: F1 verified end to end on real hardware, before 60 cards carrying a code format are committed to paper. `F11-03` is broader — the whole app at M3 — and gates nothing physical. It must still be done before send-out, because F11 is on the must-ship-before-send-out side of [07-roadmap](../../07-roadmap.md), but the print run does not wait for it.

If `F11-03` happens to run *before* printing, it subsumes `E-OPS-03`'s walkthrough — the login path is the first third of this script. Tick both; do not run the same walk twice for the sake of two checkboxes.

## Instructions

1. Test against the **deployed** site over HTTPS through the real reverse proxy, never `pnpm dev` and never an IP on the LAN. The cookie is `Secure`, the security headers come through the proxy, and the SPA fallback is a proxy concern — none of that is exercised by a dev server.
2. At least one full run on **mobile data, not wifi**: a tunnel, a lift, a handover from LTE to wifi mid-save. This is what `NetworkError` exists for, and the wrong behaviour here — logging somebody out because the train went underground — is the single most expensive bug in the product.
3. Device matrix. Cover, at minimum:
   - An iPhone on Safari. The one that breaks things: `100dvh`, the automatic zoom on inputs below 16px, its own date and select controls, and its own idea of what `SameSite=Lax` means on a redirect.
   - A mid-range Android on Chrome, the majority device.
   - One genuinely old or low-end phone from the family — a small screen, a slow CPU, an outdated browser. Borrow it at the next family gathering; that is also the moment to watch somebody use it.
   - One device with the OS font scale turned up and one with a dark system theme, since there is no dark mode and the site must not be rendered unreadable by a browser's own inversion.
   - Desktop Firefox as a secondary sanity check, because a few guests will open the link on a laptop.
4. The script, run in full on each device, in this order:
   1. Scan the QR from a **physically printed proof** (`E-OPS-04`), not from a screen. The camera reading the real ink at the real size is part of what is under test.
   2. Type the code from the card. Watch what the keyboard does: capitalisation, autocorrect, the dash somebody types out of habit.
   3. Confirm the household on the "bist das du?" screen.
   4. Read every content page. Check the bottom bar reaches all of them and that nothing needs horizontal scrolling.
   5. Fill the RSVP: a household-level scope, then a per-member override, catering fields appearing and disappearing with scope, an allergy note, transport seats, a household note.
   6. Add a plus-one, and remove it again.
   7. Save. Confirm the summary is unmissable and says what was actually submitted.
   8. Close the browser entirely, reopen the site cold, and confirm the session is still there and the answers are shown.
   9. Change one answer and save again.
5. Record evidence as you go: a filled row per device per step, pass or fail, with a screenshot for anything visual and **the request ID** for anything that errored. An ID recorded in the moment is the difference between a fixable report and "it did something weird on my mum's phone".
6. Watch somebody who is not you do steps 2 to 7 at least once, without helping. Where they hesitate is a finding, even when nothing is broken — this whole product rests on a 75-year-old getting through it alone.
7. File each finding against the story that owns the screen, with the device and the step. Anything not fixed before send-out goes into root `TODO.md` with a reason.

**Rejected: an emulator or a device farm.** BrowserStack would cover more browsers for less effort, and it cannot reproduce any of the things this story exists to find — a cracked screen at 150% font scale, a thumb that misses a 40px target, an LTE handover mid-request, or the pause before somebody works out where to type the code. The devices are physically available at the next family lunch, which is the cheaper and better test.

## Test plan

- [ ] The full script passes on the iPhone/Safari device.
- [ ] The full script passes on the mid-range Android/Chrome device.
- [ ] The full script passes on the old/low-end device.
- [ ] One full run over mobile data, including a deliberate tunnel or aeroplane-mode interruption during a save — the app reports a connection problem and does not log anybody out.
- [ ] Session survives a cold browser restart on every device.
- [ ] The printed QR scans first try, from paper, in ordinary indoor light.
- [ ] One unassisted run by somebody who has not seen the site, with the hesitations written down.
- [ ] Every finding is filed against an owning story or recorded in `TODO.md`.

## Done when

- [ ] The device matrix is filled in, with evidence, and lives in the repo rather than in a chat message.
- [ ] No open finding on the guest journey is unrecorded.
- [ ] Checkbox ticked in `README.md`.
