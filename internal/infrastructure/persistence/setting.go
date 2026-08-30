package persistence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// app_setting keys. Named constants because the same strings appear in migration
// 0001's seed insert, and a typo in one of them would read as "setting missing"
// at startup rather than as a typo.
const (
	settingRSVPDeadline         = "rsvp_deadline"
	settingDefaultAdditionLimit = "default_addition_limit"
	settingSeatingPublished     = "seating_published"
	settingGalleryVisible       = "gallery_visible"
	settingUploadsOpen          = "uploads_open"
)

// SettingStore reads the app_setting table.
//
// Read-only for now: the values are seeded by migration 0001 and changed by hand
// in sqlite3 when a gate has to open. An admin UI for them is not a story anybody
// has written, and three toggles do not justify one.
type SettingStore struct {
	database *configuration.Database
}

func NewSettingStore(database *configuration.Database) *SettingStore {
	return &SettingStore{database: database}
}

// Load reads every setting and parses it into domain.Settings.
//
// One query for the whole table rather than a lookup per key: there are five rows,
// and reading them together is what lets a missing key be reported as a missing
// key instead of as a zero value that silently opens a gate.
//
// A missing or unparseable setting is an error, not a default. These rows are
// seeded by the migration, so their absence means the database was edited by hand
// into a state the application cannot reason about — and defaulting
// `seating_published` to false would hide that, while defaulting the deadline to
// the zero time would close the RSVP form for everyone.
func (store *SettingStore) Load(ctx context.Context) (domain.Settings, error) {
	const selectSettings = `SELECT key, value FROM app_setting`

	rows, err := store.database.Read.QueryxContext(ctx, selectSettings)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("reading settings: %w", err)
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return domain.Settings{}, fmt.Errorf("scanning setting: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return domain.Settings{}, fmt.Errorf("reading settings: %w", err)
	}

	return parseSettings(values)
}

func parseSettings(values map[string]string) (domain.Settings, error) {
	var settings domain.Settings
	var err error

	if settings.RSVPDeadline, err = settingTimestamp(values, settingRSVPDeadline); err != nil {
		return domain.Settings{}, err
	}
	if settings.DefaultAdditionLimit, err = settingInt(values, settingDefaultAdditionLimit); err != nil {
		return domain.Settings{}, err
	}
	if settings.SeatingPublished, err = settingBool(values, settingSeatingPublished); err != nil {
		return domain.Settings{}, err
	}
	if settings.GalleryVisible, err = settingBool(values, settingGalleryVisible); err != nil {
		return domain.Settings{}, err
	}
	if settings.UploadsOpen, err = settingBool(values, settingUploadsOpen); err != nil {
		return domain.Settings{}, err
	}
	return settings, nil
}

func settingValue(values map[string]string, key string) (string, error) {
	value, isPresent := values[key]
	if !isPresent {
		return "", fmt.Errorf("app_setting %q is missing", key)
	}
	return value, nil
}

func settingBool(values map[string]string, key string) (bool, error) {
	raw, err := settingValue(values, key)
	if err != nil {
		return false, err
	}

	// The column holds the strings 'true' and 'false' so that SELECT * is readable
	// by hand, which is the entire point of this table.
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("app_setting %q is not a boolean: %q", key, raw)
	}
	return parsed, nil
}

func settingInt(values map[string]string, key string) (int, error) {
	raw, err := settingValue(values, key)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("app_setting %q is not a number: %q", key, raw)
	}
	return parsed, nil
}

func settingTimestamp(values map[string]string, key string) (time.Time, error) {
	raw, err := settingValue(values, key)
	if err != nil {
		return time.Time{}, err
	}
	return parseTimestamp(raw)
}
