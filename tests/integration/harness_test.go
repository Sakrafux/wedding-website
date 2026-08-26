package integration

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
)

// newTestDatabase opens a real database on a fresh temp file, closed on cleanup.
//
// A real file rather than :memory: — an in-memory database gives every connection
// its own private schema, which would make the two pools invisible to each other
// and hide exactly the bugs these tests exist to catch. The Config carries only
// DatabasePath because that is all OpenDatabase reads.
//
// E0-11 turns this into the shared harness, adding the migrations and fixtures.
func newTestDatabase(t *testing.T) *configuration.Database {
	t.Helper()

	database, err := configuration.OpenDatabase(configuration.Config{
		DatabasePath: filepath.Join(t.TempDir(), "wedding-test.db"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	// Same order as main: migrate before anything serves a request.
	require.NoError(t, persistence.Migrate(context.Background(), database.Write, slog.New(slog.NewTextHandler(io.Discard, nil))))

	return database
}

// newTestServer starts the production router on a random port.
//
// The logger is set to Error so a passing test run stays readable; a failing
// assertion says more than a request log line would.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return newTestServerWithDatabase(t, newTestDatabase(t))
}

// newTestServerWithDatabase is newTestServer for a test that needs to reach the
// database directly — closing it, or seeding it before the first request.
func newTestServerWithDatabase(t *testing.T, database *configuration.Database) *httptest.Server {
	t.Helper()

	logger := httplog.NewLogger("wedding-test", httplog.Options{LogLevel: slog.LevelError})
	server := httptest.NewServer(web.NewRouter(logger, database))
	t.Cleanup(server.Close)

	return server
}

// get performs a GET and returns status, content type and trimmed body.
func get(t *testing.T, baseURL, path string) (int, string, string) {
	t.Helper()

	response, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading body of GET %s: %v", path, err)
	}

	return response.StatusCode, response.Header.Get("Content-Type"), strings.TrimSpace(string(body))
}
