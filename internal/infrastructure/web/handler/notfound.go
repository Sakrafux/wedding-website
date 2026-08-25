package handler

import (
	"net/http"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// APINotFound answers an unmatched path under /api.
//
// Outside /api the static handler serves index.html so that SPA deep links work.
// Inside /api a miss is a genuine miss and must answer JSON — the frontend's fetch
// layer should never have to parse HTML to learn that it called the wrong URL.
func APINotFound(w http.ResponseWriter, r *http.Request) {
	httpio.WriteError(w, r, http.StatusNotFound, "not_found", "Diese Adresse gibt es nicht.")
}

// APIMethodNotAllowed answers a known /api path called with the wrong method.
func APIMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpio.WriteError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "Diese Anfrage ist hier nicht erlaubt.")
}
