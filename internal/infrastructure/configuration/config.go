package configuration

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

// Environment variable names. Collected here so the set is greppable and the
// .env.example file has an obvious counterpart in code.
const (
	envPort                = "PORT"
	envDatabasePath        = "DB_PATH"
	envPhotoDir            = "PHOTO_DIR"
	envAdminUser           = "ADMIN_USER"
	envAdminPassword       = "ADMIN_PASSWORD"
	envSessionCookieSecure = "SESSION_COOKIE_SECURE"
	envPublicBasePath      = "PUBLIC_BASE_PATH"
	envTrustedProxyCIDRs   = "TRUSTED_PROXY_CIDRS"
	envLogLevel            = "LOG_LEVEL"
)

const (
	defaultPort     = 8080
	defaultLogLevel = slog.LevelInfo

	// minAdminPasswordLength is a cheap guard against a placeholder like "admin"
	// reaching the real server. The admin password is stored in plaintext in the
	// environment, so its only protection is being hard to guess.
	minAdminPasswordLength = 12
)

// Config is the fully validated runtime configuration. Every field is set — there
// is no "unset" state to check for downstream, because Load either returns a
// complete Config or an error.
type Config struct {
	Port                int
	DatabasePath        string
	PhotoDir            string
	AdminUser           string
	AdminPassword       string
	SessionCookieSecure bool
	// PublicBasePath is the path prefix the app is reached under from the outside,
	// with a leading and no trailing slash ("/hochzeit"), or "/" when it is served
	// at the root. The reverse proxy strips the prefix, so no request ever
	// carries it and it cannot be derived — it has to be configured.
	//
	// Its one job is the session cookie's Path: with Path=/ the cookie would also be
	// sent to every other app on the same hostname. F1-B02 is what reads it; it is
	// configured here already because E0-12 is where the prefix became a fact.
	PublicBasePath string
	// TrustedProxyCIDRs lists the networks whose X-Forwarded-For header is believed.
	// Empty means "trust nobody": the direct peer address is the client IP.
	TrustedProxyCIDRs []netip.Prefix
	LogLevel          slog.Level
}

// Load reads the configuration from the process environment and validates it.
//
// All problems are collected and returned as one joined error rather than failing
// on the first: fixing environment variables one container restart at a time is
// miserable, and the operator deploying this has no debugger.
func Load() (Config, error) {
	var problems []error

	config := Config{
		DatabasePath:  requireEnv(envDatabasePath, &problems),
		PhotoDir:      requireEnv(envPhotoDir, &problems),
		AdminUser:     requireEnv(envAdminUser, &problems),
		AdminPassword: requireEnv(envAdminPassword, &problems),
	}

	if config.AdminPassword != "" && len(config.AdminPassword) < minAdminPasswordLength {
		problems = append(problems, fmt.Errorf("%s must be at least %d characters", envAdminPassword, minAdminPasswordLength))
	}

	config.Port = optionalPort(&problems)
	config.LogLevel = optionalLogLevel(&problems)
	// Defaults to true: serving the session cookie without Secure has to be a
	// deliberate act for local development, never the consequence of a forgotten
	// variable in production.
	config.SessionCookieSecure = optionalBool(envSessionCookieSecure, true, &problems)
	config.PublicBasePath = optionalBasePath(&problems)
	config.TrustedProxyCIDRs = optionalPrefixes(&problems)

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration: %w", errors.Join(problems...))
	}
	return config, nil
}

// ListenAddr returns the address for http.Server. Bound on all interfaces because
// the process runs in a container and is reached through the host's reverse proxy.
func (config Config) ListenAddr() string {
	return ":" + strconv.Itoa(config.Port)
}

// LogValue implements slog.LogValuer so that logging the config — normal at startup
// — cannot leak ADMIN_PASSWORD into the host's log collector.
func (config Config) LogValue() slog.Value {
	proxies := make([]string, 0, len(config.TrustedProxyCIDRs))
	for _, prefix := range config.TrustedProxyCIDRs {
		proxies = append(proxies, prefix.String())
	}

	return slog.GroupValue(
		slog.Int("port", config.Port),
		slog.String("db_path", config.DatabasePath),
		slog.String("photo_dir", config.PhotoDir),
		slog.String("admin_user", config.AdminUser),
		slog.String("admin_password", redacted),
		slog.Bool("session_cookie_secure", config.SessionCookieSecure),
		slog.String("public_base_path", config.PublicBasePath),
		slog.String("trusted_proxy_cidrs", strings.Join(proxies, ",")),
		slog.String("log_level", config.LogLevel.String()),
	)
}

// String exists for the same reason as LogValue: any accidental %v or fmt.Println
// of a Config must not print the password.
func (config Config) String() string {
	return config.LogValue().String()
}

const redacted = "[redacted]"

// requireEnv returns the trimmed value of a mandatory variable, recording a problem
// if it is absent or blank. Whitespace-only counts as absent: a variable set to " "
// in a compose file is a mistake, not a value.
func requireEnv(name string, problems *[]error) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		*problems = append(*problems, fmt.Errorf("%s is required", name))
	}
	return value
}

func optionalPort(problems *[]error) int {
	raw, ok := lookupEnv(envPort)
	if !ok {
		return defaultPort
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		*problems = append(*problems, fmt.Errorf("%s must be a number between 1 and 65535, got %q", envPort, raw))
		return defaultPort
	}
	return port
}

func optionalLogLevel(problems *[]error) slog.Level {
	raw, ok := lookupEnv(envLogLevel)
	if !ok {
		return defaultLogLevel
	}

	var level slog.Level
	// UnmarshalText accepts the canonical slog names ("DEBUG", "info", "WARN+2").
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		*problems = append(*problems, fmt.Errorf("%s must be one of debug, info, warn, error, got %q", envLogLevel, raw))
		return defaultLogLevel
	}
	return level
}

// optionalBasePath normalizes PUBLIC_BASE_PATH to a leading slash and no trailing
// one, so that callers can concatenate without thinking about it. "/" stays "/",
// since a cookie Path of "" is not the same thing.
//
// A value that is not a path at all fails startup rather than being repaired: the
// session cookie's scope is a security boundary, and quietly widening it to "/"
// would hand the cookie to every neighbouring app on the same hostname.
func optionalBasePath(problems *[]error) string {
	raw, ok := lookupEnv(envPublicBasePath)
	if !ok {
		return "/"
	}

	if !strings.HasPrefix(raw, "/") {
		*problems = append(*problems, fmt.Errorf("%s must start with a slash, got %q", envPublicBasePath, raw))
		return "/"
	}

	trimmed := strings.TrimRight(raw, "/")
	if trimmed == "" {
		return "/"
	}
	return trimmed
}

func optionalBool(name string, fallback bool, problems *[]error) bool {
	raw, ok := lookupEnv(name)
	if !ok {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		*problems = append(*problems, fmt.Errorf("%s must be a boolean (true/false/1/0), got %q", name, raw))
		return fallback
	}
	return value
}

// optionalPrefixes parses the comma-separated CIDR list. A malformed entry fails
// startup instead of being skipped: a silently dropped proxy network makes the
// login rate limit bypassable, and nothing later in the system would complain.
func optionalPrefixes(problems *[]error) []netip.Prefix {
	raw, ok := lookupEnv(envTrustedProxyCIDRs)
	if !ok {
		return nil
	}

	var prefixes []netip.Prefix
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			*problems = append(*problems, fmt.Errorf("%s contains an invalid CIDR %q", envTrustedProxyCIDRs, entry))
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

// lookupEnv reports a variable as absent when it is set but blank, so that an empty
// value in a compose file falls back to the default rather than failing to parse.
func lookupEnv(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	return value, value != ""
}
