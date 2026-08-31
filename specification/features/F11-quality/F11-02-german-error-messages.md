# `F11-02` — German error messages end to end, request ID surfaced

**Epic:** F11 — Cross-cutting quality · **Layer:** backend + frontend · **Depends on:** `E0-06`, `F3-B03`, `F4-B02`, `F5-F03`

## Story

As a guest whose save just failed, I want a German sentence that tells me what to do next and a number I can read out on the phone, so that the evening ends with an answer rather than with a shrug at a screen.

## Scope

**In:**

- An audit that every `domain.ErrorCode` has an entry in `errorResponses` (`internal/infrastructure/web/httpio/respond.go`), and a test that keeps it that way.
- An audit of every German sentence in that table and in `validationMessage` against the copy register in [05-design](../../05-design.md): informal "du", short, **names the fix and not the fault**.
- The request ID surfaced wherever a guest meets a failure, not only in the route error boundary.
- A check that the frontend never invents a message of its own — beyond `NetworkError`, which has no envelope to take one from.
- Resolving the stale forward references left in error-handling code.

**Out:**

- New error codes. This story audits what exists; a rule that needs a code adds it in its own story.
- What gets *logged* → [06-privacy-security](../../06-privacy-security.md) owns the logging rules, and `F12-03` owns the scrub check.
- The `/kontakt` page's presentation of the numbers → `F2-F06`. This story only cares that the fallback sentence carries a real number rather than the placeholder.

## Instructions

1. Close the gap between `domain`'s code list and `httpio`'s message table. Today an unmapped code falls through to the generic 500, which is *safe* — nothing leaks — but it is also silent to a guest and only visible in a log line nobody is reading at 21:00 on a Sunday. Export the full set of codes from `domain` and drive a table test over it, so a code added without a sentence fails the build rather than degrading in production.
2. Read the whole of `errorResponses` and `validationMessage` out loud, in order. That is the audit: the table exists precisely so that every sentence a guest can be shown fits on one screen. Check each against the register — no "Ungültige Eingabe", no "Fehler aufgetreten", no jargon ("Session", "Login", "Token", "RSVP"), one voice throughout.
3. Every rule that ships by M3 must produce a sentence that reads as an instruction. The generic `default` case in `validationMessage` is the one to hunt: any rule regularly reachable through the UI that lands on "Bitte prüfe dieses Feld." wants either its own tag case or an endpoint-level message, in the shape of `AgeValidationError`.
4. Surface the request ID beyond `RouteError`. A failed form submit renders the API's sentence inline and, today, drops the ID on the floor — which is the exact failure a guest phones about, since a route load that fails is usually just a bad connection. Show the ID under the message for any non-validation `ApiError`, in `small`, with the existing `shellLabels.requestId`. **Not** for a validation failure: the guest can fix that themselves, and a reference number next to "Bitte prüfe die markierten Felder" reads as though something broke.
5. `NetworkError` carries no ID, and that is correct — there was no response, so there is no log line to correlate with. Its sentence is the one piece of German the frontend owns; make sure it stays the only one, and say so where it is defined.
6. Verify the envelope survives the real reverse proxy. A Caddy error page or a gateway timeout is HTML, not the API's JSON, and the client maps that to `NetworkError` — which is the right answer, but it must be *observed* against the deployed proxy, not assumed from reading `client.ts`.
7. Resolve the forward references, per the rule in `CLAUDE.md`: `tests/integration/error_envelope_test.go` still says the first per-field rules "arrive with `F3-B03`" — that is `F3-B03`'s to clear when it ships, and if it has survived to here, this story deletes it. Run `grep -rn "F3-B03" --include='*.go' .` and leave nothing behind.
8. The login screen's "Klappt es nicht? Ruf uns an: …" fallback is the last escape hatch a guest has. The number is **+43 650 9408100** (decided 2026-08-31); replace the `contactPhoneNumber` placeholder in `web/src/lib/labels.ts` with it. **One** number in that sentence, not both: a guest whose code just failed twice needs somebody to ring, not a choice. The second number lives on `/kontakt` (`F2-F06`).

## Test plan

- [ ] Unit: every value in `domain`'s error-code list has an entry in `errorResponses`. This is the test that makes the audit permanent.
- [ ] Unit: every message in `errorResponses` and every branch of `validationMessage` is non-empty and free of the jargon deny-list ("Session", "Login", "Token", "RSVP", "Error", "invalid").
- [ ] Integration: an unmapped code answers 500 with the generic sentence and logs the code — the fallback still works, and is still visibly a bug.
- [ ] Integration: every failure response carries a `request_id`, and it matches the `X-Request-Id` header.
- [ ] Component: a failed form submit renders the API's sentence verbatim plus the request ID; a validation failure renders per-field messages and **no** ID.
- [ ] Component: an offline submit shows `NetworkError`'s German sentence, no ID, and does not log the guest out.
- [ ] Negative: no rendered error state anywhere contains "undefined", "[object Object]", or an English word from the deny-list.
- [ ] Unit: `contactPhoneNumber` is the real number, and the login fallback renders it.
- [ ] Manual: a 502 forced at the real proxy produces the German connection message, not a white page.

## Done when

- [ ] Every sentence a guest can be shown has been read and is in one register.
- [ ] A new error code without a German message fails `go test ./...`.
- [ ] A guest looking at any failure can read out a number that finds the request in the log.
- [ ] No `F3-B03` forward reference survives in the codebase.
- [ ] Checkbox ticked in `README.md`.
