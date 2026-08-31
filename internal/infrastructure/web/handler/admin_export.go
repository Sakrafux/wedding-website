package handler

import (
	"net/http"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/csvio"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// AdminExport serves the two CSV downloads.
//
// codes.csv is the most sensitive artefact this application produces: it is the whole
// key list leaving the server. Nothing here can enforce that it is deleted again
// after the cards are printed (E-OPS-07), so the warning lives at the point of
// download in the UI (F5-F03) and the fact of the export lives in the log below.
type AdminExport struct {
	exports *application.Exports
}

func NewAdminExport(exports *application.Exports) *AdminExport {
	return &AdminExport{exports: exports}
}

// Codes answers GET /api/admin/export/codes.csv with one row per household.
//
// German column headers, uniquely in this application, because a print shop reads
// them. The code is written in exactly the form that must appear on the card: the six
// stored characters, ungrouped, with no separator for anyone to get wrong.
func (handler *AdminExport) Codes(w http.ResponseWriter, r *http.Request) {
	households, err := handler.exports.Codes(r.Context())
	if err != nil {
		// Before Begin, so this is the one failure in the file that can still be
		// reported to the client properly.
		httpio.RespondError(w, r, err)
		return
	}

	writer := csvio.Begin(w, "codes.csv")

	if err := writer.WriteRow("haushalt", "code"); err != nil {
		logTruncatedExport(r, "codes.csv", err)
		return
	}
	for _, household := range households {
		if err := writer.WriteRow(household.DisplayName, household.Code); err != nil {
			logTruncatedExport(r, "codes.csv", err)
			return
		}
	}

	handler.finish(r, writer, "codes.csv")
}

// Guests answers GET /api/admin/export/guests.csv with one row per guest.
//
// A dump of the guest table joined onto its household, soft-deleted rows included —
// not a curated subset. This file is the release valve named in 07-roadmap: if
// send-out gets tight, F6 is dropped and this is what the caterer's numbers are read
// out of. The point of a release valve is that it does not require anyone to have
// guessed in advance which field would be wanted.
func (handler *AdminExport) Guests(w http.ResponseWriter, r *http.Request) {
	writer := csvio.Begin(w, "guests.csv")

	if err := writer.WriteRow(handler.exports.GuestExportColumns()...); err != nil {
		logTruncatedExport(r, "guests.csv", err)
		return
	}

	err := handler.exports.StreamGuests(r.Context(), writer.WriteValues)
	if err != nil {
		// The header is already on the wire, so there is no envelope to answer with:
		// the file ends short and the log is the only trace. A truncated CSV is
		// visible to whoever downloaded it, which is the best available outcome.
		logTruncatedExport(r, "guests.csv", err)
		return
	}

	handler.finish(r, writer, "guests.csv")
}

// finish flushes the response and records the export.
//
// An **info log line** with the row count, not an audit_log row. audit_log.action has
// no `read` value, and adding one is a migration that would also invite every future
// read into that table — the CHECK constraint is what keeps it a record of *changes*.
// The code list leaving the server still deserves a trace, and this is it.
func (handler *AdminExport) finish(r *http.Request, writer *csvio.Writer, filename string) {
	if err := writer.Finish(); err != nil {
		logTruncatedExport(r, filename, err)
		return
	}

	httplog.LogEntry(r.Context()).Info("csv export written", "file", filename, "rows", writer.Rows())
}

func logTruncatedExport(r *http.Request, filename string, err error) {
	httplog.LogEntry(r.Context()).Error("csv export truncated", "file", filename, "error", err)
}
