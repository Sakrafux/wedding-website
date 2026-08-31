# `F1-B01` — Code generation and normalisation

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `E0-05`

## Story

As an admin, I want household codes generated from an unambiguous alphabet and normalised on input, so that a guest typing what they see on the card always lands on their own household.

## Scope

**In:**

- `internal/domain/code.go`: `GenerateCode()`, `NormalizeCode(string)`, `FormatCode(string)`.
- Pure functions, no database, no HTTP.

**Out:**

- Persisting codes and collision retry → `F5-B03`.
- The login endpoint → `F1-B04`.

## Instructions

1. Alphabet: `23456789ABCDEFGHJKLMNPQRSTUVWXYZ` — 32 characters. Exactly four are missing: `0`, `O`, `1`, `I`. Both members of each confusable pair are dropped, never one of them, since keeping either still leaves the guest choosing between two glyphs. `L`, `U` and `V` stay — dropping them would take the alphabet below 32 and cost the modulo property in point 3. Declare it as a named constant with a comment explaining the exclusions, because the omissions look like typos to a future reader.
2. Length 6. Generate with `crypto/rand`, never `math/rand`.
3. Take bytes modulo 32 — with a 32-character alphabet this is exact, so there is no modulo bias. Note that in a comment: it is only true because 32 divides 256, and it would silently stop being true if the alphabet ever changed length.
4. `NormalizeCode`: uppercase, strip whitespace (including non-breaking spaces, which is what a paste from a PDF produces) and dashes. Return the canonical stored form.
5. No display formatting. The code is printed, exported and shown as the six stored characters. A `FormatCode` inserting a dash after character 3 existed and was removed on 2026-08-31: grouping buys nothing at this length, and it obliged the input field to decide what to do with a dash the guest typed.
6. **Do not** map lookalikes on input (`0`→`O`, `1`→`I`). Tempting, but it widens the accepted set and the alphabet already makes the confusion impossible. Record this as a rejected option.
7. Validate shape separately: after normalisation, a code is 6 characters, all from the alphabet. Return a typed error so the handler can distinguish "malformed" from "unknown".

## Test plan

- [ ] Unit: generated codes are 6 chars, all from the alphabet, over many iterations.
- [ ] Unit: no generated code ever contains `0`, `O`, `1` or `I`.
- [ ] Unit: 10,000 generated codes show no obvious skew across the alphabet.
- [ ] Unit: normalisation table — `abc-234`, `ABC 234`, ` abc234 `, `ABC–234` (en dash), `abc 234` all → `ABC234`.
- [ ] Unit: `NormalizeCode` output passes `ValidateCode` for each of the above.
- [ ] Unit: malformed inputs (5 chars, 7 chars, contains `0`) return the shape error, not a lookup attempt.

## Done when

- [ ] The domain package has no imports beyond stdlib.
- [ ] Checkbox ticked in `README.md`.
