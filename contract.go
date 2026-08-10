package cfgo

import (
	"iter"
	"time"
)

// Config defines the configuration interface
type Config interface {
	// Get retrieves a configuration value by key, or nil if absent
	Get(key string) any

	// Lookup retrieves a configuration value by key and reports whether it
	// exists, distinguishing "absent" from "present but empty"
	Lookup(key string) (any, bool)

	// GetString retrieves a string configuration value
	GetString(key string) string

	// GetInt retrieves an integer configuration value
	GetInt(key string) int

	// GetInt64 retrieves an int64 configuration value
	GetInt64(key string) int64

	// GetFloat64 retrieves a float64 configuration value
	GetFloat64(key string) float64

	// GetBool retrieves a boolean configuration value
	GetBool(key string) bool

	// GetDuration retrieves a time.Duration configuration value
	GetDuration(key string) time.Duration

	// GetStringSlice retrieves a comma-separated string slice configuration value
	GetStringSlice(key string) []string

	// GetStringMap retrieves all values under "key." as a map
	GetStringMap(key string) map[string]any

	// Each iterates over a point-in-time snapshot of all configuration
	// values; it is safe to mutate the config while iterating
	Each() iter.Seq2[string, any]

	// Set sets a configuration value; Set values override every source and
	// survive Reload
	Set(key string, value any)

	// Has checks if a configuration key exists
	Has(key string) bool

	// All returns a copy of all configuration values. Note that by default
	// this includes every process environment variable — treat the result as
	// sensitive and do not log it wholesale.
	All() map[string]any

	// Reload rebuilds the configuration from files, environment, and
	// sources; on error the previous configuration is left intact
	Reload() error

	// AddSource registers a configuration source; call Reload to load it
	AddSource(source Source)
}

// Source is the minimal contract for a custom configuration source. Values
// it returns override env files and system environment variables.
type Source interface {
	// Load loads configuration from the source
	Load() (map[string]any, error)
}

// NamedSource is an optional extension of Source. When implemented, the name
// is used in Reload error messages to identify the failing source.
type NamedSource interface {
	Source

	// Name returns the name of the configuration source
	Name() string
}

// ConfigSource is the pre-v1.2 source contract.
//
// Deprecated: implement Source (and optionally NamedSource) instead. cfgo
// has never called Watch. Existing ConfigSource implementations keep working
// with AddSource because ConfigSource satisfies Source.
type ConfigSource interface {
	// Name returns the name of the configuration source
	Name() string

	// Load loads configuration from the source
	Load() (map[string]any, error)

	// Watch watches for configuration changes
	Watch(callback func(map[string]any)) error
}
