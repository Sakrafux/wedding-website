package handler

import (
	"net/http"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// System serves the operational endpoints: liveness, readiness and the JSON
// fallbacks for unmatched routes under /api.
type System struct {
	database *configuration.Database
}

// NewSystem builds the operational handlers over the database handles.
func NewSystem(database *configuration.Database) *System {
	return &System{database: database}
}

// Health reports that the process is up and serving.
//
// Unauthenticated by design and staying that way: the reverse proxy and any
// container restart check poll it. It deliberately reports nothing about the
// database, the photo volume or the build — a health check that fails on a
// subsystem outage causes restart loops instead of surfacing the real problem,
// and restarting fixes neither an unmounted volume nor a wrong DB_PATH. Use Ready
// for that. An unauthenticated endpoint should also leak no version information.
func (system *System) Health(w http.ResponseWriter, r *http.Request) {
	httpio.WriteJSON(w, r, http.StatusOK, dto.HealthResponse{Status: "ok"})
}

// Ready reports whether the dependencies are usable: 200 when both database pools
// respond, 503 otherwise.
//
// This is the post-deploy smoke check — the one request that proves the binary
// really opened the database it was configured with, rather than serving happily
// from a file it created inside the container. The failure reason is logged and
// never returned, because the endpoint is unauthenticated and a driver error
// carries the database path.
func (system *System) Ready(w http.ResponseWriter, r *http.Request) {
	if err := system.database.Ping(r.Context()); err != nil {
		httplog.LogEntry(r.Context()).Error("readiness check failed", "error", err)
		httpio.RespondError(w, r, httpio.ErrNotReady)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, dto.ReadyResponse{Status: "ok", Database: "ok"})
}

// APINotFound answers an unmatched path under /api.
//
// Outside /api the static handler serves index.html so that SPA deep links work.
// Inside /api a miss is a genuine miss and must answer JSON — the frontend's fetch
// layer should never have to parse HTML to learn that it called the wrong URL.
func (system *System) APINotFound(w http.ResponseWriter, r *http.Request) {
	httpio.RespondError(w, r, httpio.ErrNotFound)
}

// APIMethodNotAllowed answers a known /api path called with the wrong method.
func (system *System) APIMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpio.RespondError(w, r, httpio.ErrMethodNotAllowed)
}
