# CFGO - Environment-Based Configuration for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/anxuanzi/cfgo.svg)](https://pkg.go.dev/github.com/anxuanzi/cfgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/anxuanzi/cfgo)](https://goreportcard.com/report/github.com/anxuanzi/cfgo)
[![License](https://img.shields.io/github/license/anxuanzi/cfgo)](LICENSE)

CFGO is a lightweight, zero-dependency configuration library for Go applications. It provides a simple and flexible way to manage configuration values from environment variables and files, with support for different environments (development, production, etc.).

## Features

- **Zero Dependencies**: Built using only the Go standard library
- **Environment-Based**: Automatically loads configuration from environment variables and `.env` files
- **Environment-Specific Configuration**: Supports different configurations for different environments
- **Type Conversion**: Easily retrieve configuration values as strings, integers, booleans, etc.
- **Extensible**: Add custom configuration sources
- **Thread-Safe**: Safe for concurrent use
- **Caching**: Caches configuration values for better performance

## Who's Using CFGO

CFGO is being used in the following projects:

- [Goe Application Development Framework](https://github.com/oeasenet/goe) - A comprehensive application development framework

## Installation

```bash
go get github.com/anxuanzi/cfgo
```

## Quick Start

### Using the Global Instance

```go
package main

import (
    "fmt"
    "github.com/anxuanzi/cfgo"
)

func main() {
    // Use the global configuration instance directly
    dbHost := cfgo.GetString("DB_HOST")
    dbPort := cfgo.GetInt("DB_PORT")
    debugMode := cfgo.GetBool("DEBUG_MODE")

    fmt.Printf("Database: %s:%d, Debug Mode: %v\n", dbHost, dbPort, debugMode)
}
```

### Using a Custom Instance

```go
package main

import (
    "fmt"
    "github.com/anxuanzi/cfgo"
)

func main() {
    // Create a new configuration instance
    config := cfgo.New()

    // Get configuration values
    dbHost := config.GetString("DB_HOST")
    dbPort := config.GetInt("DB_PORT")
    debugMode := config.GetBool("DEBUG_MODE")

    fmt.Printf("Database: %s:%d, Debug Mode: %v\n", dbHost, dbPort, debugMode)
}
```

## Environment Files

CFGO automatically loads configuration from the following files (in order of precedence):

1. `.env` - Base configuration file
2. `.local.env` - Local overrides (not committed to version control)
3. `.{APP_ENV}.env` - Environment-specific configuration (e.g., `.dev.env`, `.prod.env`)

Example `.env` file:

```
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret

# Application Settings
DEBUG_MODE=true
LOG_LEVEL=debug
API_TIMEOUT=30s
```

## Environment Variables

System environment variables take precedence over values defined in environment files. This allows you to override configuration values at runtime without modifying files.

## Usage Examples

### Basic Usage

#### Using the Global Instance

```go
// Get string value
appName := cfgo.GetString("APP_NAME")

// Get integer value
port := cfgo.GetInt("PORT")

// Get boolean value
debug := cfgo.GetBool("DEBUG")

// Get duration value
timeout := cfgo.GetDuration("TIMEOUT")

// Check if a configuration key exists
if cfgo.Has("FEATURE_FLAG") {
    // Use feature flag
}
```

#### Using a Custom Instance

```go
// Create a new configuration instance
config := cfgo.New()

// Get string value
appName := config.GetString("APP_NAME")

// Get integer value
port := config.GetInt("PORT")

// Get boolean value
debug := config.GetBool("DEBUG")

// Get duration value
timeout := config.GetDuration("TIMEOUT")

// Check if a configuration key exists
if config.Has("FEATURE_FLAG") {
    // Use feature flag
}
```

### Setting Values Programmatically

#### Using the Global Instance

```go
// Set a configuration value
cfgo.Set("CACHE_TTL", "60s")

// Get the value
cacheTTL := cfgo.GetDuration("CACHE_TTL")
```

#### Using a Custom Instance

```go
config := cfgo.New()

// Set a configuration value
config.Set("CACHE_TTL", "60s")

// Get the value
cacheTTL := config.GetDuration("CACHE_TTL")
```

### Working with Slices and Maps

#### Using the Global Instance

```go
// Get a comma-separated list as a slice
// ENV: ALLOWED_ORIGINS=https://example.com,https://api.example.com
origins := cfgo.GetStringSlice("ALLOWED_ORIGINS")

// Get a nested map
// ENV: DB_CONFIG.HOST=localhost
// ENV: DB_CONFIG.PORT=5432
dbConfig := cfgo.GetStringMap("DB_CONFIG")
```

#### Using a Custom Instance

```go
config := cfgo.New()

// Get a comma-separated list as a slice
// ENV: ALLOWED_ORIGINS=https://example.com,https://api.example.com
origins := config.GetStringSlice("ALLOWED_ORIGINS")

// Get a nested map
// ENV: DB_CONFIG.HOST=localhost
// ENV: DB_CONFIG.PORT=5432
dbConfig := config.GetStringMap("DB_CONFIG")
```

### Custom Configuration Sources

#### Using the Global Instance

```go
type JSONConfigSource struct {
    path string
    data map[string]any
}

func NewJSONConfigSource(path string) *JSONConfigSource {
    return &JSONConfigSource{path: path}
}

func (j *JSONConfigSource) Name() string {
    return "json-config"
}

func (j *JSONConfigSource) Load() (map[string]any, error) {
    // Load JSON file and parse it
    // ...
    return j.data, nil
}

func (j *JSONConfigSource) Watch(callback func(map[string]any)) error {
    // Implement file watching if needed
    return nil
}

// Usage with global instance
cfgo.AddSource(NewJSONConfigSource("config.json"))
cfgo.Reload()
```

#### Using a Custom Instance

```go
type JSONConfigSource struct {
    path string
    data map[string]any
}

func NewJSONConfigSource(path string) *JSONConfigSource {
    return &JSONConfigSource{path: path}
}

func (j *JSONConfigSource) Name() string {
    return "json-config"
}

func (j *JSONConfigSource) Load() (map[string]any, error) {
    // Load JSON file and parse it
    // ...
    return j.data, nil
}

func (j *JSONConfigSource) Watch(callback func(map[string]any)) error {
    // Implement file watching if needed
    return nil
}

// Usage with custom instance
config := cfgo.New()
config.AddSource(NewJSONConfigSource("config.json"))
config.Reload()
```

### Reloading Configuration

#### Using the Global Instance

```go
// Later, reload configuration (e.g., after receiving SIGHUP)
err := cfgo.Reload()
if err != nil {
    log.Fatalf("Failed to reload configuration: %v", err)
}
```

#### Using a Custom Instance

```go
config := cfgo.New()

// Later, reload configuration (e.g., after receiving SIGHUP)
err := config.Reload()
if err != nil {
    log.Fatalf("Failed to reload configuration: %v", err)
}
```

## API Reference

### Config Interface

```go
type Config interface {
    // Get retrieves a configuration value by key
    Get(key string) any

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

    // GetStringSlice retrieves a string slice configuration value
    GetStringSlice(key string) []string

    // GetStringMap retrieves a string map configuration value
    GetStringMap(key string) map[string]any

    // Set sets a configuration value
    Set(key string, value any)

    // Has checks if a configuration key exists
    Has(key string) bool

    // All returns all configuration values
    All() map[string]any

    // Reload reloads the configuration from sources
    Reload() error

    // AddSource adds a configuration source
    AddSource(source ConfigSource)
}
```

### ConfigSource Interface

```go
type ConfigSource interface {
    // Name returns the name of the configuration source
    Name() string

    // Load loads configuration from the source
    Load() (map[string]any, error)

    // Watch watches for configuration changes
    Watch(callback func(map[string]any)) error
}
```

## Best Practices

1. **Use Environment Variables for Sensitive Information**: Avoid storing sensitive information like API keys and passwords in files.

2. **Use `.local.env` for Local Development**: Keep a `.local.env` file for local development settings and add it to `.gitignore`.

3. **Set Default Values**: Always check if a configuration value exists and provide sensible defaults.

4. **Use Environment-Specific Files**: Create separate configuration files for different environments (`.dev.env`, `.prod.env`, etc.).

5. **Validate Configuration**: Validate critical configuration values at startup to fail fast if something is misconfigured.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
