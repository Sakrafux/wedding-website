package middleware

import (
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// Recoverer turns a panic into a logged stack trace and a generic 500 in the
// standard error envelope.
//
// It replaces chi's Recoverer, which writes an empty body — leaving the frontend's
// single error path with nothing to parse — and prints the stack to stderr instead
// of into the structured log. Here the trace is logged through httplog, so it
// arrives with the request ID and the frontend still gets a real envelope.
//
// A panic is a bug, not a runtime condition, so the response says nothing about it:
// the stack trace, the Go type and the panic value never leave the process.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			// http.ErrAbortHandler is the standard library's way of aborting a
			// response on purpose; swallowing it would turn a deliberate abort into
			// a 500. Re-panic and let net/http handle it, as chi's Recoverer does.
			if recovered == http.ErrAbortHandler {
				panic(recovered)
			}

			httplog.LogEntry(r.Context()).Error("panic recovered",
				"panic", recovered,
				"stack", string(debug.Stack()),
			)

			// The status line may already be on the wire if the handler panicked
			// after writing. Appending an envelope there would corrupt the body
			// the client is already reading, so the log is the only record.
			if written, ok := w.(chimiddleware.WrapResponseWriter); ok && written.Status() != 0 {
				return
			}

			// ErrInternal rather than the recovered value: the stack is already logged
			// above, and passing the panic on would log it a second time.
			httpio.RespondError(w, r, httpio.ErrInternal)
		}()

		next.ServeHTTP(w, r)
	})
}
