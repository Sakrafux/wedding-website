import type { Address } from "./labels";

/**
 * Singular or plural, from the people we seeded into the household.
 *
 * `origin === "seeded"` and not the current member count: a guest invited alone
 * who adds a companion is still the person we wrote to, and switching to "ihr"
 * the moment they name somebody reads as the site correcting them (`F2-F08`).
 */
export function addressFor(members: readonly { origin: string }[]): Address {
  return members.filter((member) => member.origin === "seeded").length > 1 ? "plural" : "singular";
}
