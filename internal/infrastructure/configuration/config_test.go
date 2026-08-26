package configuration_test

import (
	"log/slog"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// validEnvironment is the minimum set that must produce a valid Config: the four
// required variables and nothing else.
func validEnvironment() map[string]string {
	return map[string]string{
		"DB_PATH":        "/data/wedding.db",
		"PHOTO_DIR":      "/data/photos",
		"ADMIN_USER":     "admin",
		"ADMIN_PASSWORD": "a-long-enough-password",
	}
}

// setEnvironment applies exactly the given variables and clears every other variable
// the loader reads, so a value leaking in from the developer's shell cannot change
// the outcome of a test.
func setEnvironment(t *testing.T, environment map[string]string) {
	t.Helper()

	known := []string{
		"PORT", "DB_PATH", "PHOTO_DIR", "ADMIN_USER", "ADMIN_PASSWORD",
		"SESSION_COOKIE_SECURE", "TRUSTED_PROXY_CIDRS", "LOG_LEVEL",
	}
	for _, name := range known {
		t.Setenv(name, "")
	}
	for name, value := range environment {
		t.Setenv(name, value)
	}
}

func TestLoadDefaults(t *testing.T) {
	setEnvironment(t, validEnvironment())

	config, err := configuration.Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, ":8080", config.ListenAddr())
	assert.Equal(t, "/data/wedding.db", config.DatabasePath)
	assert.Equal(t, "/data/photos", config.PhotoDir)
	assert.Equal(t, "admin", config.AdminUser)
	assert.Equal(t, "a-long-enough-password", config.AdminPassword)
	assert.True(t, config.SessionCookieSecure, "must default to true")
	assert.Empty(t, config.TrustedProxyCIDRs)
	assert.Equal(t, slog.LevelInfo, config.LogLevel)
}

func TestLoadAllVariablesSet(t *testing.T) {
	environment := validEnvironment()
	environment["PORT"] = "9000"
	environment["SESSION_COOKIE_SECURE"] = "false"
	environment["TRUSTED_PROXY_CIDRS"] = "172.18.0.0/16, ::1/128"
	environment["LOG_LEVEL"] = "debug"
	setEnvironment(t, environment)

	config, err := configuration.Load()
	require.NoError(t, err)

	assert.Equal(t, 9000, config.Port)
	assert.False(t, config.SessionCookieSecure)
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("172.18.0.0/16"),
		netip.MustParsePrefix("::1/128"),
	}, config.TrustedProxyCIDRs)
	assert.Equal(t, slog.LevelDebug, config.LogLevel)
}

func TestLoadMissingRequiredVariable(t *testing.T) {
	for _, name := range []string{"DB_PATH", "PHOTO_DIR", "ADMIN_USER", "ADMIN_PASSWORD"} {
		t.Run(name, func(t *testing.T) {
			environment := validEnvironment()
			delete(environment, name)
			setEnvironment(t, environment)

			_, err := configuration.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), name)
		})
	}
}

func TestLoadReportsAllProblemsAtOnce(t *testing.T) {
	environment := validEnvironment()
	delete(environment, "DB_PATH")
	delete(environment, "ADMIN_USER")
	environment["PORT"] = "not-a-port"
	setEnvironment(t, environment)

	_, err := configuration.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_PATH")
	assert.Contains(t, err.Error(), "ADMIN_USER")
	assert.Contains(t, err.Error(), "PORT")
}

func TestLoadRejectsShortAdminPassword(t *testing.T) {
	environment := validEnvironment()
	environment["ADMIN_PASSWORD"] = "admin"
	setEnvironment(t, environment)

	_, err := configuration.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ADMIN_PASSWORD")
}

func TestLoadRejectsMalformedValues(t *testing.T) {
	cases := map[string]struct {
		name  string
		value string
	}{
		"cidr":          {name: "TRUSTED_PROXY_CIDRS", value: "172.18.0.0/16,nonsense"},
		"bare ip":       {name: "TRUSTED_PROXY_CIDRS", value: "172.18.0.1"},
		"port range":    {name: "PORT", value: "70000"},
		"cookie secure": {name: "SESSION_COOKIE_SECURE", value: "yes-please"},
		"log level":     {name: "LOG_LEVEL", value: "chatty"},
	}

	for description, testCase := range cases {
		t.Run(description, func(t *testing.T) {
			environment := validEnvironment()
			environment[testCase.name] = testCase.value
			setEnvironment(t, environment)

			_, err := configuration.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.name)
		})
	}
}

// TestLogRepresentationRedactsPassword guards the one mistake in this package that
// would be invisible: the startup log line is written on every boot, so a leak here
// ends up in every log archive.
func TestLogRepresentationRedactsPassword(t *testing.T) {
	const password = "correct-horse-battery"

	environment := validEnvironment()
	environment["ADMIN_PASSWORD"] = password
	setEnvironment(t, environment)

	config, err := configuration.Load()
	require.NoError(t, err)

	assert.NotContains(t, config.String(), password)
	assert.NotContains(t, config.LogValue().String(), password)
	assert.Contains(t, config.String(), "[redacted]")
}
