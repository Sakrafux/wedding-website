package persistence

import (
	"fmt"
	"time"
)

// timestampLayout is the one format every TEXT timestamp column uses: RFC3339 in
// UTC, second precision. It is exactly what the schema's
// `strftime('%Y-%m-%dT%H:%M:%SZ', 'now')` defaults produce, so a row written by
// hand in sqlite3 and a row written by the app are indistinguishable — and because
// the format is fixed-width and UTC, lexicographic comparison in SQL is
// chronological comparison, which is what makes the expiry queries below plain
// string comparisons.
const timestampLayout = time.RFC3339

// formatTimestamp renders a time for storage. The UTC conversion is not optional:
// a local-zone timestamp would still be valid RFC3339 but would sort wrongly
// against every other row.
func formatTimestamp(at time.Time) string {
	return at.UTC().Format(timestampLayout)
}

// parseTimestamp reads a stored timestamp back.
//
// A parse failure is returned rather than swallowed as a zero time: a zero time
// reads as "expired in the year 1" everywhere downstream, which would log a
// confused guest out and give no hint why.
func parseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(timestampLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing timestamp %q: %w", value, err)
	}
	return parsed.UTC(), nil
}
