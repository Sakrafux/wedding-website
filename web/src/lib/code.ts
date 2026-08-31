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
 * Dashes are stripped although the card no longer prints one: six characters are
 * short enough to check against a card ungrouped, and dropping the group separator
 * removed every awkward case this field used to have — a dash that vanished under
 * the cursor, or a caret that jumped because we inserted one. Guests still type
 * dashes out of habit and a word processor still turns a hyphen into an en dash, so
 * all three are accepted and none survive.
 *
 * The length cap is applied last, after any dash is gone: capping the raw input
 * would truncate a typed ABC-234 — seven characters — to ABC-23 and quietly cost
 * the guest the final digit of their code.
 */
export function normalizeCode(input: string): string {
  return input
    .replace(/[\s\p{Pd}]/gu, "")
    .toUpperCase()
    .slice(0, codeLength);
}
