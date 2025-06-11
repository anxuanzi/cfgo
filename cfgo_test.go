package cfgo

import (
	"os"
	"testing"
	"time"
)

// MockConfigSource implements ConfigSource for testing
type MockConfigSource struct {
	name string
	data map[string]any
}

func NewMockConfigSource(name string, data map[string]any) *MockConfigSource {
	return &MockConfigSource{
		name: name,
		data: data,
	}
}

func (m *MockConfigSource) Name() string {
	return m.name
}

func (m *MockConfigSource) Load() (map[string]any, error) {
	return m.data, nil
}

func (m *MockConfigSource) Watch(callback func(map[string]any)) error {
	// Not implemented for tests
	return nil
}

// Helper function to create a temporary env file for testing
func createTempEnvFile(t *testing.T, filename string, content string) {
	t.Helper()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}
}

// Helper function to clean up temporary env files
func cleanupTempEnvFiles(t *testing.T, files ...string) {
	t.Helper()
	for _, file := range files {
		os.Remove(file)
	}
}

func TestNew(t *testing.T) {
	// Set environment variable for testing
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	cfg := New()

	// Test that environment variables are loaded
	if cfg.GetString("TEST_ENV_VAR") != "test_value" {
		t.Errorf("Expected TEST_ENV_VAR to be 'test_value', got '%s'", cfg.GetString("TEST_ENV_VAR"))
	}
}

func TestLoadEnvFiles(t *testing.T) {
	// Create temporary env files
	createTempEnvFile(t, ".env", "BASE_KEY=base_value\nSHARED_KEY=base_shared")
	createTempEnvFile(t, ".local.env", "LOCAL_KEY=local_value\nSHARED_KEY=local_shared")
	createTempEnvFile(t, ".dev.env", "DEV_KEY=dev_value\nSHARED_KEY=dev_shared")
	defer cleanupTempEnvFiles(t, ".env", ".local.env", ".dev.env")

	// Set APP_ENV to dev
	os.Setenv("APP_ENV", "dev")
	defer os.Unsetenv("APP_ENV")

	cfg := New()

	// Test that values from all files are loaded with correct precedence
	if cfg.GetString("BASE_KEY") != "base_value" {
		t.Errorf("Expected BASE_KEY to be 'base_value', got '%s'", cfg.GetString("BASE_KEY"))
	}

	if cfg.GetString("LOCAL_KEY") != "local_value" {
		t.Errorf("Expected LOCAL_KEY to be 'local_value', got '%s'", cfg.GetString("LOCAL_KEY"))
	}

	if cfg.GetString("DEV_KEY") != "dev_value" {
		t.Errorf("Expected DEV_KEY to be 'dev_value', got '%s'", cfg.GetString("DEV_KEY"))
	}

	// Test that shared key has the value from the highest precedence file (.dev.env)
	if cfg.GetString("SHARED_KEY") != "dev_shared" {
		t.Errorf("Expected SHARED_KEY to be 'dev_shared', got '%s'", cfg.GetString("SHARED_KEY"))
	}
}

func TestGetMethods(t *testing.T) {
	c := &config{
		data:  make(map[string]any),
		cache: make(map[string]any),
	}

	// Set test values
	c.data["string_key"] = "test_string"
	c.data["int_key"] = "42"
	c.data["int64_key"] = "9223372036854775807"
	c.data["float_key"] = "3.14159"
	c.data["bool_key_true"] = "true"
	c.data["bool_key_false"] = "false"
	c.data["duration_key"] = "1h30m"
	c.data["slice_key"] = "item1, item2, item3"
	c.data["map_prefix.key1"] = "value1"
	c.data["map_prefix.key2"] = "value2"
	c.data["empty_key"] = ""

	// Test GetString
	if c.GetString("string_key") != "test_string" {
		t.Errorf("GetString failed, expected 'test_string', got '%s'", c.GetString("string_key"))
	}
	if c.GetString("nonexistent_key") != "" {
		t.Errorf("GetString for nonexistent key should return empty string, got '%s'", c.GetString("nonexistent_key"))
	}

	// Test GetInt
	if c.GetInt("int_key") != 42 {
		t.Errorf("GetInt failed, expected 42, got %d", c.GetInt("int_key"))
	}
	if c.GetInt("nonexistent_key") != 0 {
		t.Errorf("GetInt for nonexistent key should return 0, got %d", c.GetInt("nonexistent_key"))
	}
	if c.GetInt("string_key") != 0 {
		t.Errorf("GetInt for non-int value should return 0, got %d", c.GetInt("string_key"))
	}

	// Test GetInt64
	if c.GetInt64("int64_key") != 9223372036854775807 {
		t.Errorf("GetInt64 failed, expected 9223372036854775807, got %d", c.GetInt64("int64_key"))
	}

	// Test GetFloat64
	if c.GetFloat64("float_key") != 3.14159 {
		t.Errorf("GetFloat64 failed, expected 3.14159, got %f", c.GetFloat64("float_key"))
	}

	// Test GetBool
	if !c.GetBool("bool_key_true") {
		t.Errorf("GetBool failed for true value")
	}
	if c.GetBool("bool_key_false") {
		t.Errorf("GetBool failed for false value")
	}

	// Test GetDuration
	expected := 90 * time.Minute
	if c.GetDuration("duration_key") != expected {
		t.Errorf("GetDuration failed, expected %v, got %v", expected, c.GetDuration("duration_key"))
	}

	// Test GetStringSlice
	expectedSlice := []string{"item1", "item2", "item3"}
	actualSlice := c.GetStringSlice("slice_key")
	if len(actualSlice) != len(expectedSlice) {
		t.Errorf("GetStringSlice failed, expected %v, got %v", expectedSlice, actualSlice)
	}
	for i, v := range expectedSlice {
		if actualSlice[i] != v {
			t.Errorf("GetStringSlice failed at index %d, expected '%s', got '%s'", i, v, actualSlice[i])
		}
	}

	// Test GetStringMap
	stringMap := c.GetStringMap("map_prefix")
	if len(stringMap) != 2 {
		t.Errorf("GetStringMap failed, expected 2 items, got %d", len(stringMap))
	}
	if stringMap["key1"] != "value1" || stringMap["key2"] != "value2" {
		t.Errorf("GetStringMap returned incorrect values: %v", stringMap)
	}

	// Test empty values
	if c.GetString("empty_key") != "" {
		t.Errorf("GetString for empty value should return empty string, got '%s'", c.GetString("empty_key"))
	}
	if c.GetInt("empty_key") != 0 {
		t.Errorf("GetInt for empty value should return 0, got %d", c.GetInt("empty_key"))
	}
	if c.GetBool("empty_key") != false {
		t.Errorf("GetBool for empty value should return false")
	}
}

func TestSetAndHas(t *testing.T) {
	c := &config{
		data:  make(map[string]any),
		cache: make(map[string]any),
	}

	// Test Has with non-existent key
	if c.Has("test_key") {
		t.Errorf("Has should return false for non-existent key")
	}

	// Test Set
	c.Set("test_key", "test_value")

	// Test Has with existent key
	if !c.Has("test_key") {
		t.Errorf("Has should return true for existent key")
	}

	// Test that value was set correctly
	if c.GetString("test_key") != "test_value" {
		t.Errorf("Set failed, expected 'test_value', got '%s'", c.GetString("test_key"))
	}

	// Test cache invalidation
	c.cache["test_key"] = "cached_value"
	c.Set("test_key", "new_value")

	// Get should return the new value, not the cached one
	if c.Get("test_key") != "new_value" {
		t.Errorf("Cache invalidation failed, expected 'new_value', got '%v'", c.Get("test_key"))
	}
}

func TestAll(t *testing.T) {
	c := &config{
		data:  make(map[string]any),
		cache: make(map[string]any),
	}

	c.data["key1"] = "value1"
	c.data["key2"] = "value2"

	all := c.All()
	if len(all) != 2 {
		t.Errorf("All failed, expected 2 items, got %d", len(all))
	}

	if all["key1"] != "value1" || all["key2"] != "value2" {
		t.Errorf("All returned incorrect values: %v", all)
	}
}

func TestReload(t *testing.T) {
	// Create temporary env files
	createTempEnvFile(t, ".env", "RELOAD_KEY=initial_value")
	defer cleanupTempEnvFiles(t, ".env")

	cfg := New()

	// Test initial value
	if cfg.GetString("RELOAD_KEY") != "initial_value" {
		t.Errorf("Expected RELOAD_KEY to be 'initial_value', got '%s'", cfg.GetString("RELOAD_KEY"))
	}

	// Update env file
	createTempEnvFile(t, ".env", "RELOAD_KEY=updated_value")

	// Reload configuration
	err := cfg.Reload()
	if err != nil {
		t.Errorf("Reload failed: %v", err)
	}

	// Test updated value
	if cfg.GetString("RELOAD_KEY") != "updated_value" {
		t.Errorf("Expected RELOAD_KEY to be 'updated_value', got '%s'", cfg.GetString("RELOAD_KEY"))
	}
}

func TestAddSource(t *testing.T) {
	cfg := New()

	// Create mock config source
	mockData := map[string]any{
		"mock_key": "mock_value",
	}
	mockSource := NewMockConfigSource("mock", mockData)

	// Add source
	cfg.AddSource(mockSource)

	// Reload to load from the source
	err := cfg.Reload()
	if err != nil {
		t.Errorf("Reload failed: %v", err)
	}

	// Test that value from mock source is loaded
	if cfg.GetString("mock_key") != "mock_value" {
		t.Errorf("Expected mock_key to be 'mock_value', got '%s'", cfg.GetString("mock_key"))
	}
}

func TestCaching(t *testing.T) {
	c := &config{
		data:  make(map[string]any),
		cache: make(map[string]any),
	}

	c.data["test_key"] = "test_value"

	// First access should cache the value
	val1 := c.Get("test_key")
	if val1 != "test_value" {
		t.Errorf("Expected 'test_value', got '%v'", val1)
	}

	// Modify the data directly (bypassing Set which would invalidate cache)
	c.data["test_key"] = "modified_value"

	// Second access should return cached value
	val2 := c.Get("test_key")
	if val2 != "test_value" {
		t.Errorf("Expected cached 'test_value', got '%v'", val2)
	}

	// Clear cache
	c.cache = make(map[string]any)

	// Third access should get the modified value
	val3 := c.Get("test_key")
	if val3 != "modified_value" {
		t.Errorf("Expected 'modified_value', got '%v'", val3)
	}
}
