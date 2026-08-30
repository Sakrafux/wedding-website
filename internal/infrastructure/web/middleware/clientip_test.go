package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

// dockerBridge is the shape of a real TRUSTED_PROXY_CIDRS value: the container
// network the reverse proxy reaches this process on.
var dockerBridge = []netip.Prefix{netip.MustParsePrefix("172.16.0.0/12")}

func requestFrom(peer, forwardedFor string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	request.RemoteAddr = peer
	if forwardedFor != "" {
		request.Header.Set(forwardedForHeader, forwardedFor)
	}
	return request
}

func TestClientIPUsesTheForwardedHeaderBehindATrustedProxy(t *testing.T) {
	t.Parallel()

	resolved := resolveClientIP(requestFrom("172.18.0.5:41234", "203.0.113.7"), dockerBridge)

	assert.Equal(t, "203.0.113.7", resolved)
}

// The header is a plain request header. Believing it from an untrusted peer would
// let anyone spend a fresh rate-limit budget on every attempt by changing it.
func TestClientIPIgnoresTheForwardedHeaderFromAnUntrustedPeer(t *testing.T) {
	t.Parallel()

	resolved := resolveClientIP(requestFrom("198.51.100.4:41234", "203.0.113.7"), dockerBridge)

	assert.Equal(t, "198.51.100.4", resolved)
}

// The misconfiguration the startup warning is about: with nothing trusted, the
// header is never read, so the limiter keys on the proxy and all guests share one
// budget. Wrong, but wrong in the safe direction and visible immediately.
func TestClientIPIgnoresTheForwardedHeaderWhenNothingIsTrusted(t *testing.T) {
	t.Parallel()

	resolved := resolveClientIP(requestFrom("172.18.0.5:41234", "203.0.113.7"), nil)

	assert.Equal(t, "172.18.0.5", resolved)
}

// Everything left of the rightmost untrusted entry was appended by whoever is
// upstream of our own proxy — which is to say by the client, which is to say by
// nobody we can believe.
func TestClientIPTakesTheRightmostUntrustedHop(t *testing.T) {
	t.Parallel()

	resolved := resolveClientIP(
		requestFrom("172.18.0.5:41234", "10.9.9.9, 203.0.113.7, 172.18.0.2"),
		dockerBridge,
	)

	assert.Equal(t, "203.0.113.7", resolved, "the trusted hop is skipped, the one before it is the client")
}

func TestClientIPFallsBackToThePeer(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"no header at all":        "",
		"header is not addresses": "unknown",
		// A chain of nothing but our own proxies says nothing about the client.
		"every hop is trusted": "172.18.0.2, 172.18.0.5",
		// Garbage in the rightmost position means the chain is not the one we
		// expect; walking further left would read entries the client wrote.
		"garbage nearest to us": "203.0.113.7, nonsense",
	}

	for name, forwardedFor := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "172.18.0.5", resolveClientIP(requestFrom("172.18.0.5:41234", forwardedFor), dockerBridge))
		})
	}
}

// A proxy on the IPv4 side of a dual-stack listener presents ::ffff:172.18.0.1,
// which no IPv4 prefix contains until it is unmapped. Getting this wrong would
// silently stop trusting the real proxy.
func TestClientIPTrustsAnIPv4MappedProxyAddress(t *testing.T) {
	t.Parallel()

	resolved := resolveClientIP(requestFrom("[::ffff:172.18.0.5]:41234", "203.0.113.7"), dockerBridge)

	assert.Equal(t, "203.0.113.7", resolved)
}

func TestClientIPFromContextIsEmptyWithoutTheMiddleware(t *testing.T) {
	t.Parallel()

	assert.Empty(t, ClientIPFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context()))
}
