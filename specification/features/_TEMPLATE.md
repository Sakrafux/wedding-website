# `<ID>` — <Short title>

**Epic:** <F3 — RSVP> · **Layer:** <backend | frontend | ops> · **Depends on:** <IDs, or "nothing">

> Copy this file, do not edit it. Delete the guidance in blockquotes as you fill it in.
>
> No `Status:` line, ever. Progress lives in `README.md` and nowhere else.

## Story

> One sentence, from the actor's point of view. Who wants what, and why. If the "why" is not obvious, the story is probably not worth doing.

As a <guest | admin | developer>, I want <capability> so that <outcome>.

## Scope

**In:**

- <the concrete things this story delivers>

**Out:**

- <the adjacent things it deliberately does not deliver, with the ID that covers them where one exists>

## Instructions

> Precise enough to implement without re-deriving decisions, but not a transcription of the code. Reference the spec instead of restating it: "per 03-data-model" beats a copied column list that will drift.

1. <step>
2. <step>

## Contract

> Backend stories only. The endpoint, the request and response DTOs, the error codes. This is what the paired frontend story is allowed to rely on — and the only thing it is allowed to rely on.

```http
<METHOD> /api/<path>
```

Request:

```json
{}
```

Response `200`:

```json
{}
```

Errors: `<code>` → `<HTTP status>` → `<German message shown to the user>`

## Test plan

> What proves this works. Name the cases, not the assertions. A story with no test plan is not ready to be built.

- [ ] <unit: the domain rule, in isolation>
- [ ] <integration: the endpoint against a real temp SQLite>
- [ ] <negative: the case a bug would silently pass>
- [ ] <privacy: no `code`, `admin_note` or budget field in any guest-facing response, where applicable>

## Done when

- [ ] <observable outcome, verifiable by someone who did not write it>
- [ ] Tests above pass; `go test ./...` is green
- [ ] Checkbox ticked in `README.md`
