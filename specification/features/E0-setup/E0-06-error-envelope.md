# `E0-06` — Error envelope, request ID, panic recovery

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-01`

## Story

As a guest, I want every failure to produce a sentence I can understand and a short code I can read out over the phone, so that being stuck is recoverable without a developer.

## Scope

**In:**

- One JSON error shape for the whole API.
- A request ID generated per request, returned in a header and embedded in every error body.
- Panic recovery that logs the stack and returns a clean 500.
- A helper for field-keyed validation errors.

**Out:**

- Validation rules themselves → the stories that own each endpoint.
- Frontend rendering → `F1-F01` onwards.

## Instructions

1. Shape, per [04-architecture](../../04-architecture.md):

   ```json
   { "error": { "code": "invalid_code", "message": "…", "request_id": "a1b2c3", "fields": { "code": "…" } } }
   ```

   `fields` is present only for validation failures.
2. `message` is **German and safe to show verbatim.** Never a Go error string, never SQL, never a file path.
3. Request ID: short and human-readable — 6–8 base32 characters, not a UUID. It gets read aloud over the phone by someone who is already flustered.
4. Return it as an `X-Request-Id` header on **all** responses, and include it in the log line so a phone call maps to a log entry.
5. Panic recovery logs the stack at error level with the request ID, and returns a generic German 500 body. A stack trace never reaches the client.
6. Central `respondError(w, r, appErr)` helper. Handlers do not hand-roll JSON error bodies — that is how one endpoint ends up leaking a database message.

## Test plan

- [ ] Integration: an unknown route returns the envelope with a `request_id`.
- [ ] Integration: `X-Request-Id` is present on success responses too.
- [ ] Integration: a handler that panics returns 500 in the envelope, and the response body contains no stack trace or Go type name.
- [ ] Integration: a validation error returns `fields` keyed by input name.
- [ ] Unit: the same request ID appears in the log line and the response body.

## Done when

- [ ] Every error path in the app produces the same shape.
- [ ] Checkbox ticked in `README.md`.
