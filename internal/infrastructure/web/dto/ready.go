package dto

// ReadyResponse is the body of GET /api/ready.
//
// Only ever "ok": a failing check answers with the error envelope instead, and the
// reason stays in the log. The endpoint is unauthenticated, so it must not report
// which dependency failed or why — a driver error carries the database path.
type ReadyResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
