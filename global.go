package cfgo

import (
	"iter"
	"sync"
	"time"
)

// defaultConfig lazily constructs the shared instance on first use, so
// importing cfgo has no side effects: no files are read and no environment
// snapshot is taken until the package-level API is actually called.
var defaultConfig = sync.OnceValue(func() Config { return New() })

// Default returns the shared Config instance used by the package-level
// functions, creating it on first call.
func Default() Config {
	return defaultConfig()
}

// Get retrieves a configuration value by key from the default config
func Get(key string) any {
	return defaultConfig().Get(key)
}

// Lookup retrieves a configuration value by key from the default config and
// reports whether it exists
func Lookup(key string) (any, bool) {
	return defaultConfig().Lookup(key)
}

// GetString retrieves a string configuration value from the default config
func GetString(key string) string {
	return defaultConfig().GetString(key)
}

// GetInt retrieves an integer configuration value from the default config
func GetInt(key string) int {
	return defaultConfig().GetInt(key)
}

// GetInt64 retrieves an int64 configuration value from the default config
func GetInt64(key string) int64 {
	return defaultConfig().GetInt64(key)
}

// GetFloat64 retrieves a float64 configuration value from the default config
func GetFloat64(key string) float64 {
	return defaultConfig().GetFloat64(key)
}

// GetBool retrieves a boolean configuration value from the default config
func GetBool(key string) bool {
	return defaultConfig().GetBool(key)
}

// GetDuration retrieves a time.Duration configuration value from the default config
func GetDuration(key string) time.Duration {
	return defaultConfig().GetDuration(key)
}

// GetStringSlice retrieves a string slice configuration value from the default config
func GetStringSlice(key string) []string {
	return defaultConfig().GetStringSlice(key)
}

// GetStringMap retrieves a string map configuration value from the default config
func GetStringMap(key string) map[string]any {
	return defaultConfig().GetStringMap(key)
}

// Each iterates over a snapshot of all values in the default config
func Each() iter.Seq2[string, any] {
	return defaultConfig().Each()
}

// Set sets a configuration value in the default config
func Set(key string, value any) {
	defaultConfig().Set(key, value)
}

// Has checks if a configuration key exists in the default config
func Has(key string) bool {
	return defaultConfig().Has(key)
}

// All returns a copy of all configuration values from the default config
func All() map[string]any {
	return defaultConfig().All()
}

// Reload reloads the default config from its sources
func Reload() error {
	return defaultConfig().Reload()
}

// AddSource adds a configuration source to the default config
func AddSource(source Source) {
	defaultConfig().AddSource(source)
}
