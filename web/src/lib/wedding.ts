/**
 * The wedding itself, as one constant.
 *
 * Everything static about the day is hardcoded (02-features), but the *date* is
 * hardcoded exactly once: the countdown, the age question and every sentence that
 * spells it out read it from here. It is still a working assumption until the venue
 * is booked — which is the whole reason it is one constant rather than a string in
 * four files.
 */

/** Local midnight on the wedding day. Month is 0-based in `Date`. */
export const weddingDate = new Date(2027, 6, 17);

/** "17. Juli 2027", for prose. */
export const weddingDateLong = new Intl.DateTimeFormat("de-DE", {
  day: "numeric",
  month: "long",
  year: "numeric",
}).format(weddingDate);

/** "17.07.2027", for the places that want it short — the age question, above all. */
export const weddingDateShort = new Intl.DateTimeFormat("de-DE", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
}).format(weddingDate);

/**
 * Whole days from today until the wedding: positive before, 0 on the day, negative
 * after.
 *
 * Both sides are taken to local midnight rather than subtracting two timestamps. A
 * guest opening the site at 23:50 must be told the same number as the person beside
 * them at 00:10 — a difference of milliseconds must not become a difference of a day.
 */
export function daysUntilWedding(now = new Date()): number {
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const millisecondsPerDay = 24 * 60 * 60 * 1000;

  return Math.round((weddingDate.getTime() - today.getTime()) / millisecondsPerDay);
}
