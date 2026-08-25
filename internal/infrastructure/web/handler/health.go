package handler

import (
	"net/http"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// Health reports that the process is up and serving.
//
// Unauthenticated by design and staying that way: the reverse proxy and any
// container restart check poll it. It deliberately reports nothing about the
// database, the photo volume or the build — a health check that fails on a
// subsystem outage causes restart loops instead of surfacing the real problem,
// and an unauthenticated endpoint should leak no version information.
//
// It is a plain function, not a method, because it has no dependencies to hold.
func Health(w http.ResponseWriter, r *http.Request) {
	httpio.WriteJSON(w, r, http.StatusOK, dto.HealthResponse{Status: "ok"})
}
