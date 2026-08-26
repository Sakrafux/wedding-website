// Command wedding is the single binary that serves the wedding web app: the JSON
// API under /api and, from E0-09 onward, the embedded React SPA on every other path.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
)

// Server timeouts. Every one of these is set explicitly because Go's zero value
// means "no timeout": a single slow or half-open client would otherwise hold a
// connection open indefinitely, and there is no load balancer in front of us that
// would notice.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	// Generous enough for a photo upload (F9) over a phone connection.
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second

	// shutdownTimeout is the grace period granted to in-flight requests after a
	// termination signal. Docker's default kill delay is 10s, so anything longer
	// would be cut short by SIGKILL anyway.
	shutdownTimeout = 10 * time.Second
)

func main() {
	// Configuration is loaded before the real logger exists, because it carries the
	// log level. A load failure is therefore reported through a plain stderr logger.
	config, err := configuration.Load()
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("configuration failed", "error", err)
		os.Exit(1)
	}

	// JSON output: logs go to stdout for the host's log collector, not to a human
	// tailing a terminal. Concise keeps chi's per-request lines to one entry.
	logger := httplog.NewLogger("wedding", httplog.Options{
		LogLevel:        config.LogLevel,
		JSON:            true,
		Concise:         true,
		TimeFieldFormat: time.RFC3339,
	})
	// Config redacts ADMIN_PASSWORD in its log representation; see Config.LogValue.
	logger.Info("configuration loaded", "config", config)

	if err := run(config, logger); err != nil {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}

// run wires logger, router and server, then blocks until the process is asked to
// stop. It exists so that no code path calls os.Exit directly, which would skip
// deferred cleanup — the database handle and photo directory will rely on that later.
func run(config configuration.Config, logger *httplog.Logger) error {
	listenAddr := config.ListenAddr()

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           web.NewRouter(logger),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// SIGTERM is what `docker stop` sends; SIGINT is Ctrl-C in development.
	signalCtx, stopListeningForSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopListeningForSignals()

	listenFailed := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", listenAddr)
		listenFailed <- server.ListenAndServe()
	}()

	select {
	case err := <-listenFailed:
		// ErrServerClosed only appears if something else already called Shutdown,
		// which is not a failure.
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signalCtx.Done():
		logger.Info("shutdown signal received, draining connections")
	}

	// Fresh context: signalCtx is already cancelled, and Shutdown needs one that lives.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	return server.Shutdown(shutdownCtx)
}
