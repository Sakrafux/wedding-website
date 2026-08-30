package middleware

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// forwardedForHeader is the header the reverse proxy appends the real client to.
const forwardedForHeader = "X-Forwarded-For"

type clientIPContextKey struct{}

// ClientIP resolves the caller's address once and puts it in the request context.
//
// X-Forwarded-For is believed only when the request actually arrived from one of
// trustedProxies. The header is a plain request header that anyone can set, so an
// unconditionally trusted one makes the login rate limit bypassable by writing a
// different value on every attempt — which is why chi's own RealIP middleware,
// which trusts it unconditionally, is deliberately not used here.
//
// With no trusted proxies configured the header is never read at all. That is the
// safe direction: over-trusting silently disables the limiter, while under-trusting
// merely lumps every guest behind the proxy onto one budget, which is visible the
// moment it happens.
func ClientIP(trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resolved := resolveClientIP(r, trustedProxies)

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientIPContextKey{}, resolved)))
		})
	}
}

// ClientIPFromContext returns the resolved client address, or the empty string if
// the ClientIP middleware did not run.
//
// Callers use this rather than r.RemoteAddr, so that "who is this request from" has
// exactly one answer and the trusted-proxy rule cannot be bypassed by a handler
// that reads the header itself.
func ClientIPFromContext(ctx context.Context) string {
	address, _ := ctx.Value(clientIPContextKey{}).(string)
	return address
}

// resolveClientIP picks the address to attribute this request to.
//
// The peer is the proxy in production and the guest in local development, so it is
// the fallback in both cases. When the peer is a trusted proxy, the rightmost entry
// of X-Forwarded-For that is not itself a trusted proxy is the closest thing to the
// real client that we can trust: everything further left was appended by whoever is
// upstream of our own proxy, which is to say the client, which is to say nobody.
func resolveClientIP(r *http.Request, trustedProxies []netip.Prefix) string {
	peer := peerAddress(r)

	if len(trustedProxies) == 0 || !isTrusted(peer, trustedProxies) {
		return peer
	}

	hops := strings.Split(r.Header.Get(forwardedForHeader), ",")
	for index := len(hops) - 1; index >= 0; index-- {
		candidate, err := netip.ParseAddr(strings.TrimSpace(hops[index]))
		if err != nil {
			// An entry that is not an address means the chain is not the one we
			// expect — an absent header, or something injected. Continuing further
			// left would walk into entries the client wrote, so stop here.
			return peer
		}
		if !isTrustedAddr(candidate, trustedProxies) {
			return candidate.String()
		}
	}

	// Every hop is a proxy we run. The peer is then the most specific truthful
	// answer available.
	return peer
}

// peerAddress is the address the connection actually came from, with the port
// stripped. Falls back to the raw value, which keeps a key that is at least stable
// per client if the format is ever something unexpected.
func peerAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isTrusted(address string, trustedProxies []netip.Prefix) bool {
	parsed, err := netip.ParseAddr(address)
	if err != nil {
		return false
	}
	return isTrustedAddr(parsed, trustedProxies)
}

func isTrustedAddr(address netip.Addr, trustedProxies []netip.Prefix) bool {
	// Unmap first: a proxy reaching us over IPv4 may present ::ffff:172.18.0.1,
	// which no IPv4 prefix would ever contain.
	address = address.Unmap()

	for _, prefix := range trustedProxies {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}
