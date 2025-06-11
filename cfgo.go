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
	mu      sync.RWMutex
	data    map[string]any
	sources []ConfigSource
	cache   map[string]any
}

// New creates a new config instance
func New() Config {
	c := &config{
		data:    make(map[string]any),
		sources: make([]ConfigSource, 0),
		cache:   make(map[string]any),
	}

	// Load default env files
	c.loadEnvFiles()

	// Load system environment variables
	c.loadSystemEnv()

	return c
}

// loadEnvFiles loads environment files in order
func (c *config) loadEnvFiles() {
	// Always load .env first if it exists
	c.loadEnvFile(".env")

	// Then load .local.env which can override .env values
	c.loadEnvFile(".local.env")

	// Load environment-specific file based on APP_ENV
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	envFile := fmt.Sprintf(".%s.env", env)
	c.loadEnvFile(envFile)
}

// loadEnvFile loads a single env file
func (c *config) loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // File doesn't exist, skip
	}
	defer file.Close()

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

		c.data[key] = value
	}
}

// loadSystemEnv loads system environment variables
func (c *config) loadSystemEnv() {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			c.data[parts[0]] = parts[1]
		}
	}
}

// Get retrieves a configuration value by key
func (c *config) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check cache first
	if val, ok := c.cache[key]; ok {
		return val
	}

	// Check data
	if val, ok := c.data[key]; ok {
		c.cache[key] = val
		return val
	}

	return nil
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
			mapKey := strings.TrimPrefix(k, prefix)
			result[mapKey] = v
		}
	}

	return result
}

// Set sets a configuration value
func (c *config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	delete(c.cache, key) // Invalidate cache
}

// Has checks if a configuration key exists
func (c *config) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.data[key]
	return ok
}

// All returns all configuration values
func (c *config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any)
	for k, v := range c.data {
		result[k] = v
	}

	return result
}

// Reload reloads the configuration from sources
func (c *config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear cache
	c.cache = make(map[string]any)

	// Clear data
	c.data = make(map[string]any)

	// Reload env files first (.env, then .local.env, then .{GOE_ENV}.env)
	c.loadEnvFiles()

	// Reload system env last to ensure highest priority
	c.loadSystemEnv()

	// Reload from custom sources
	for _, source := range c.sources {
		data, err := source.Load()
		if err != nil {
			return err
		}

		for k, v := range data {
			c.data[k] = v
		}
	}

	return nil
}

// AddSource adds a configuration source
func (c *config) AddSource(source ConfigSource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append(c.sources, source)
}
