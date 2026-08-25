package dto

// HealthResponse is the body of GET /api/health.
//
// Omits version, uptime and any subsystem status on purpose: the endpoint is
// unauthenticated, so it must reveal nothing about the deployment.
type HealthResponse struct {
	Status string `json:"status"`
}
