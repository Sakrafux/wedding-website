/**
 * The household login code, on the client side.
 *
 * The server is the authority — domain.NormalizeCode and domain.ValidateCode decide
 * what is accepted — and this exists only so the field behaves predictably and the
 * value that gets sent is the value the guest meant. Nothing here rejects anything.
 */

/** Matches domain.codeLength on the server. */
export const codeLength = 6;

/**
 * sanitizeCodeInput is what the field displays while the guest types.
 *
 * It keeps a dash the guest typed or pasted, because the code is printed as
 * ABC-234 and a dash that vanishes under the cursor reads as "the app is not taking
 * my input" — someone then assumes they mistyped and starts over. What it does not
 * do is *insert* one: auto-formatting means moving the caret on every keystroke, and
 * the classic symptom is backspace jumping to the end of the field, which is a small
 * annoyance for most people and an unrecoverable one for the guests this screen is
 * designed around. So the dash is the guest's to type, never ours.
 *
 * At most one dash, never leading, and never once six characters are in — those are
 * the positions where a dash cannot be part of the printed form and is therefore a
 * stray keypress. Whitespace is dropped outright, since it is invisible and no code
 * contains any.
 *
 * The cap counts code characters only, so a pasted ABC-234 — seven characters —
 * survives intact instead of losing its final digit.
 */
export function sanitizeCodeInput(input: string): string {
  let sanitized = "";
  let characters = 0;
  let hasDash = false;

  for (const character of input.toUpperCase()) {
    if (/\p{Pd}/u.test(character)) {
      if (!hasDash && characters > 0 && characters < codeLength) {
        sanitized += "-";
        hasDash = true;
      }
      continue;
    }
    if (/\s/u.test(character)) {
      continue;
    }
    if (characters === codeLength) {
      continue;
    }
    sanitized += character;
    characters += 1;
  }

  return sanitized;
}

/**
 * normalizeCode mirrors domain.NormalizeCode: upper case, no whitespace, no dashes.
 *
 * This is the form that goes over the wire, so what the guest sees in the field and
 * what the API is asked about may differ by exactly one dash. Applied on submit
 * rather than on every keystroke — see sanitizeCodeInput for why the displayed value
 * is left alone.
 *
 * The length cap is applied last, after the dash is gone: capping the raw input
 * would truncate a pasted ABC-234 — seven characters — to ABC-23 and quietly cost
 * the guest the final digit of their code.
 */
export function normalizeCode(input: string): string {
  return input
    .replace(/[\s\p{Pd}]/gu, "")
    .toUpperCase()
    .slice(0, codeLength);
}
