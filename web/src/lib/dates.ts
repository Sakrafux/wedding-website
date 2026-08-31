/**
 * Dates as this application shows them: German, short, and absolute.
 *
 * The strings here are produced by `Intl` rather than written out, which is the one
 * exception to "every German string lives in labels.ts" — a month name from a
 * translation table would be a second German locale to maintain alongside the one
 * the browser already has.
 */

/** Fixed locale: the whole product is German, and the admin's OS locale is not a
    reason for a date to render differently than it does for everyone else. */
const locale = "de-DE";

const shortDate = new Intl.DateTimeFormat(locale, { day: "2-digit", month: "2-digit", year: "numeric" });
const relative = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });

/** Days within which a relative hint ("vor 3 Tagen") is worth showing. Beyond a
    month it stops being easier to read than the date itself. */
const relativeHintDays = 30;

const millisecondsPerDay = 24 * 60 * 60 * 1000;

/** formatShortDate renders an RFC3339 timestamp as `03.11.2026`. */
export function formatShortDate(timestamp: string): string {
  return shortDate.format(new Date(timestamp));
}

/**
 * formatRelativeDays renders "heute", "gestern", "vor 3 Tagen" — or null once the
 * date is old enough that the absolute one reads better.
 *
 * Both forms are shown together on purpose: the absolute date is what gets read out
 * on the phone, and the relative one is what makes a column scannable.
 */
export function formatRelativeDays(timestamp: string, now = new Date()): string | null {
  const days = Math.round((new Date(timestamp).getTime() - now.getTime()) / millisecondsPerDay);

  if (Math.abs(days) > relativeHintDays) {
    return null;
  }
  return relative.format(days, "day");
}
