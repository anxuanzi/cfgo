package cfgo

import (
	"bufio"
	"fmt"
	"os"
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
	sources   []ConfigSource
}

// Global instance of the configuration
var defaultConfig Config

// init initializes the default global configuration instance
func init() {
	defaultConfig = New()
}

// New creates a new config instance
func New() Config {
	data := make(map[string]any)
	loadEnvFiles(data)
	loadSystemEnv(data)

	return &config{
		data:      data,
		overrides: make(map[string]any),
		sources:   make([]ConfigSource, 0),
	}
}

// loadEnvFiles loads environment files into m, later files overriding earlier
// ones: .env (base) < .{APP_ENV}.env (environment-specific) < .local.env
// (personal, uncommitted overrides).
func loadEnvFiles(m map[string]any) {
	// Base configuration
	loadEnvFile(m, ".env")

	// Environment-specific file based on APP_ENV (default "dev")
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	loadEnvFile(m, fmt.Sprintf(".%s.env", env))

	// Personal overrides load last so they win over committed files
	loadEnvFile(m, ".local.env")
}

// loadEnvFile loads a single env file into m
func loadEnvFile(m map[string]any, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // File doesn't exist, skip
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		m[key] = value
	}
}

// loadSystemEnv loads system environment variables into m
func loadSystemEnv(m map[string]any) {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
}

// Get retrieves a configuration value by key
func (c *config) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if val, ok := c.overrides[key]; ok {
		return val
	}
	return c.data[key]
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

	// Split by comma
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
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
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.overrides[key]; ok {
		return true
	}
	_, ok := c.data[key]
	return ok
}

// All returns all configuration values
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
// registered sources. The rebuild is all-or-nothing: if any source fails, the
// previous configuration stays fully intact and the error names the source.
// Values set via Set are kept.
func (c *config) Reload() error {
	c.mu.RLock()
	sources := make([]ConfigSource, len(c.sources))
	copy(sources, c.sources)
	c.mu.RUnlock()

	// Build the new state without holding the lock; readers keep working
	// against the previous state until the swap.
	newData := make(map[string]any)
	loadEnvFiles(newData)
	loadSystemEnv(newData)

	for _, source := range sources {
		data, err := source.Load()
		if err != nil {
			return fmt.Errorf("cfgo: reload source %q: %w", source.Name(), err)
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

// AddSource adds a configuration source
func (c *config) AddSource(source ConfigSource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append(c.sources, source)
}

// Global functions that delegate to the default config instance

// Get retrieves a configuration value by key from the default config
func Get(key string) any {
	return defaultConfig.Get(key)
}

// GetString retrieves a string configuration value from the default config
func GetString(key string) string {
	return defaultConfig.GetString(key)
}

// GetInt retrieves an integer configuration value from the default config
func GetInt(key string) int {
	return defaultConfig.GetInt(key)
}

// GetInt64 retrieves an int64 configuration value from the default config
func GetInt64(key string) int64 {
	return defaultConfig.GetInt64(key)
}

// GetFloat64 retrieves a float64 configuration value from the default config
func GetFloat64(key string) float64 {
	return defaultConfig.GetFloat64(key)
}

// GetBool retrieves a boolean configuration value from the default config
func GetBool(key string) bool {
	return defaultConfig.GetBool(key)
}

// GetDuration retrieves a time.Duration configuration value from the default config
func GetDuration(key string) time.Duration {
	return defaultConfig.GetDuration(key)
}

// GetStringSlice retrieves a string slice configuration value from the default config
func GetStringSlice(key string) []string {
	return defaultConfig.GetStringSlice(key)
}

// GetStringMap retrieves a string map configuration value from the default config
func GetStringMap(key string) map[string]any {
	return defaultConfig.GetStringMap(key)
}

// Set sets a configuration value in the default config
func Set(key string, value any) {
	defaultConfig.Set(key, value)
}

// Has checks if a configuration key exists in the default config
func Has(key string) bool {
	return defaultConfig.Has(key)
}

// All returns all configuration values from the default config
func All() map[string]any {
	return defaultConfig.All()
}

// Reload reloads the configuration from sources in the default config
func Reload() error {
	return defaultConfig.Reload()
}

// AddSource adds a configuration source to the default config
func AddSource(source ConfigSource) {
	defaultConfig.AddSource(source)
}
