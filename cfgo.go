package cfgo

import (
	"fmt"
	"iter"
	"strconv"
	"strings"
	"sync"
	"time"
)

// config implements the Config interface
type config struct {
	mu        sync.RWMutex
	data      map[string]any // values loaded from files, environment, and sources
	overrides map[string]any // values set programmatically via Set; survive Reload
	sources   []Source
	opts      options
}

// New creates a new config instance. With no options it loads the default
// env files from the current working directory and then the process
// environment; options customize the directory, file set, environment
// selector variable, and whether the process environment is included.
func New(opts ...Option) Config {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	data := make(map[string]any)
	_ = loadEnvFiles(data, o) // best-effort at construction; Reload surfaces errors
	if o.includeSystemEnv {
		loadSystemEnv(data)
	}

	return &config{
		data:      data,
		overrides: make(map[string]any),
		sources:   make([]Source, 0),
		opts:      o,
	}
}

// Get retrieves a configuration value by key
func (c *config) Get(key string) any {
	val, _ := c.Lookup(key)
	return val
}

// Lookup retrieves a configuration value by key and reports whether it exists
func (c *config) Lookup(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if val, ok := c.overrides[key]; ok {
		return val, true
	}
	val, ok := c.data[key]
	return val, ok
}

// GetString retrieves a string configuration value
func (c *config) GetString(key string) string {
	val := c.Get(key)
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt retrieves an integer configuration value
func (c *config) GetInt(key string) int {
	val := c.GetString(key)
	if val == "" {
		return 0
	}

	i, _ := strconv.Atoi(val)
	return i
}

// GetInt64 retrieves an int64 configuration value
func (c *config) GetInt64(key string) int64 {
	val := c.GetString(key)
	if val == "" {
		return 0
	}

	i, _ := strconv.ParseInt(val, 10, 64)
	return i
}

// GetFloat64 retrieves a float64 configuration value
func (c *config) GetFloat64(key string) float64 {
	val := c.GetString(key)
	if val == "" {
		return 0
	}

	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// GetBool retrieves a boolean configuration value
func (c *config) GetBool(key string) bool {
	val := c.GetString(key)
	if val == "" {
		return false
	}

	b, _ := strconv.ParseBool(val)
	return b
}

// GetDuration retrieves a time.Duration configuration value
func (c *config) GetDuration(key string) time.Duration {
	val := c.GetString(key)
	if val == "" {
		return 0
	}

	d, _ := time.ParseDuration(val)
	return d
}

// GetStringSlice retrieves a string slice configuration value
func (c *config) GetStringSlice(key string) []string {
	val := c.GetString(key)
	if val == "" {
		return []string{}
	}

	return splitList(val)
}

// splitList splits a comma-separated value into trimmed, non-empty items.
func splitList(val string) []string {
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetStringMap retrieves a string map configuration value
func (c *config) GetStringMap(key string) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any)
	prefix := key + "."

	for k, v := range c.data {
		if strings.HasPrefix(k, prefix) {
			result[strings.TrimPrefix(k, prefix)] = v
		}
	}
	for k, v := range c.overrides {
		if strings.HasPrefix(k, prefix) {
			result[strings.TrimPrefix(k, prefix)] = v
		}
	}

	return result
}

// Each iterates over a point-in-time snapshot of all configuration values.
// The snapshot is taken when iteration starts, so it is safe to call Set or
// Reload from inside the loop.
func (c *config) Each() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range c.All() {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Set sets a configuration value. Values set this way take precedence over
// every loaded source and survive Reload.
func (c *config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.overrides == nil {
		c.overrides = make(map[string]any)
	}
	c.overrides[key] = value
}

// Has checks if a configuration key exists
func (c *config) Has(key string) bool {
	_, ok := c.Lookup(key)
	return ok
}

// All returns a copy of all configuration values
func (c *config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any, len(c.data)+len(c.overrides))
	for k, v := range c.data {
		result[k] = v
	}
	for k, v := range c.overrides {
		result[k] = v
	}

	return result
}

// Reload rebuilds the configuration from env files, system environment, and
// registered sources, honoring the options the instance was created with.
// The rebuild is all-or-nothing: if anything fails, the previous
// configuration stays fully intact and the error identifies the source.
// Values set via Set are kept.
func (c *config) Reload() error {
	c.mu.RLock()
	o := c.opts
	sources := make([]Source, len(c.sources))
	copy(sources, c.sources)
	c.mu.RUnlock()

	// Build the new state without holding the lock; readers keep working
	// against the previous state until the swap.
	newData := make(map[string]any)
	if err := loadEnvFiles(newData, o); err != nil {
		return err
	}
	if o.includeSystemEnv {
		loadSystemEnv(newData)
	}

	for _, source := range sources {
		data, err := source.Load()
		if err != nil {
			return fmt.Errorf("cfgo: reload source %q: %w", sourceName(source), err)
		}
		for k, v := range data {
			newData[k] = v
		}
	}

	c.mu.Lock()
	c.data = newData
	c.mu.Unlock()

	return nil
}

// sourceName identifies a source for error messages, preferring an explicit
// Name over the source's type.
func sourceName(s Source) string {
	if n, ok := s.(NamedSource); ok {
		return n.Name()
	}
	return fmt.Sprintf("%T", s)
}

// AddSource registers a configuration source. Sources are consulted on
// Reload, in registration order, and override env files and system
// environment variables; call Reload after adding sources to load them.
func (c *config) AddSource(source Source) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append(c.sources, source)
}
