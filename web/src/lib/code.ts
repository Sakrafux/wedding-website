/**
 * The household login code, on the client side.
 *
 * The server is the authority — domain.NormalizeCode and domain.ValidateCode decide
 * what is accepted — and this exists only so the field shows the guest the value
 * that will actually be sent. Nothing here rejects anything.
 */

/** Matches domain.codeLength on the server. */
export const codeLength = 6;

/**
 * normalizeCode mirrors domain.NormalizeCode: upper case, no whitespace, no dashes.
 *
 * Applied as the guest types, so the value in the field is the value that gets
 * sent and nobody has to trust that something invisible happens on submit.
 *
 * The dash from the printed ABC-234 is stripped rather than re-inserted for
 * display. Auto-formatting would mean moving the caret on every keystroke, and the
 * classic symptom is backspace jumping to the end of the field — a small annoyance
 * for most people and an unrecoverable one for the guests this screen is designed
 * around. The hint text carries the printed form instead.
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
