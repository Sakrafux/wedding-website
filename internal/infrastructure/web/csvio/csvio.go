// Package csvio writes the CSV downloads the admin area offers.
//
// It exists so the encoding decisions below are made once and cannot diverge
// between the file the print shop gets and the file we read ourselves. It is its own
// package next to httpio for the same reason that one is: response writing is not a
// handler's business, and a helper living in handler would be copied the second time
// it was needed.
package csvio

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// The encoding, and why it is not RFC 4180.
//
// **UTF-8 with a BOM, semicolon-delimited, CRLF.** This is a decision about Excel,
// not about correctness. German Excel splits on the locale's list separator, which is
// `;`, so a comma-delimited file lands entirely in column A; and without a BOM it
// reads the bytes as Latin-1, so `Müller` becomes `MÃ¼ller`. Both failures are
// silent, both would be found by the print shop rather than by us, and one of them
// ends up on eighty invitation cards.
//
// The cost is real and accepted: the files are then not RFC 4180 and not what
// `pandas.read_csv` expects by default.
const (
	fieldSeparator = ";"
	lineSeparator  = "\r\n"
	// The BOM as an escape rather than as the character itself: a literal U+FEFF in
	// a Go source file is a compile error, and it would be invisible in review anyway.
	byteOrderMark = "\ufeff"
)

// Writer streams CSV rows to a response.
//
// Streaming rather than building the file in memory. Not for the sixty rows this
// application has — for the habit, and because the alternative sets a precedent for
// the photo ZIP in F10-B03, where it would matter.
type Writer struct {
	buffered *bufio.Writer
	rows     int
}

// Begin sends the download headers and the byte-order mark, and returns the writer
// for the rows.
//
// Content-Disposition is `attachment`: a CSV that renders in the browser instead of
// downloading is a CSV somebody copies out of the page by hand.
//
// `Cache-Control: no-store` for two reasons. codes.csv is the whole key list leaving
// the server, and a copy of it sitting in a browser or proxy cache is a copy nobody
// knows about. It is also the only thing that makes re-downloading trustworthy: these
// URLs never change, so without it a stale export is served again after the data — or
// the format — has moved on, which is a genuinely confusing way to lose an afternoon.
//
// Nothing can be reported to the client after this point — the status line is
// already on the wire — so the caller logs a later failure and leaves the response
// truncated. That is the same trade httpio.WriteJSON makes, and for the same reason.
func Begin(w http.ResponseWriter, filename string) *Writer {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	buffered := bufio.NewWriter(w)
	_, _ = buffered.WriteString(byteOrderMark)

	return &Writer{buffered: buffered}
}

// WriteRow writes one row of strings, typically the header.
func (writer *Writer) WriteRow(fields ...string) error {
	encoded := make([]string, 0, len(fields))
	for _, field := range fields {
		encoded = append(encoded, quoteIfNeeded(field))
	}

	if _, err := writer.buffered.WriteString(strings.Join(encoded, fieldSeparator) + lineSeparator); err != nil {
		return fmt.Errorf("writing csv row: %w", err)
	}
	writer.rows++
	return nil
}

// WriteValues writes one row of database values, rendering each as text.
//
// The rendering lives here rather than in the store because it is a decision about
// the file: a NULL is an empty field, and an integer flag stays 0/1 rather than
// becoming "true", so that a column in the export reads exactly as the column in the
// database does.
func (writer *Writer) WriteValues(values []any) error {
	fields := make([]string, 0, len(values))
	for _, value := range values {
		fields = append(fields, render(value))
	}
	return writer.WriteRow(fields...)
}

// Rows is how many rows have been written, header included. Used for the log line
// that records an export having happened.
func (writer *Writer) Rows() int {
	return writer.rows
}

// Finish flushes the buffer. A failure here means the client went away mid-download,
// which the caller logs.
func (writer *Writer) Finish() error {
	if err := writer.buffered.Flush(); err != nil {
		return fmt.Errorf("flushing csv: %w", err)
	}
	return nil
}

// quoteIfNeeded wraps a field only when leaving it bare would break the row.
//
// Quoting everything was the earlier decision and was wrong in practice: every
// value a person copies out of a cell — a login code, a name — comes with quotes
// attached, which Excel shows but does not strip on copy. So the quotes are back to
// doing the one job they exist for: a field carrying the separator, a quote or a
// line break still splits a row, and that is not hypothetical in a file of
// free-text notes.
//
// A doubled quote is the CSV escape for a literal one.
func quoteIfNeeded(field string) string {
	if !strings.ContainsAny(field, fieldSeparator+`"`+"\r\n") {
		return field
	}
	return `"` + strings.ReplaceAll(field, `"`, `""`) + `"`
}

// render turns a scanned database value into a field.
func render(value any) string {
	switch typed := value.(type) {
	case nil:
		// NULL is an empty field and not the word "null": the file is read in a
		// spreadsheet, where an empty cell is what "we do not know" looks like.
		return ""
	case string:
		return typed
	case []byte:
		// SQLite hands back BLOB and, depending on the column's declared type,
		// sometimes TEXT as bytes.
		return string(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}
