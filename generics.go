package cfgo

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ErrNotFound is wrapped by GetAs when the requested key does not exist.
var ErrNotFound = errors.New("key not found")

// Parseable enumerates the types the generic getters can produce from a
// configuration value.
type Parseable interface {
	string | int | int64 | float64 | bool | time.Duration | []string
}

// GetAs retrieves the value for key from c converted to T. Unlike the
// zero-value getters (GetInt, GetBool, ...), it distinguishes the three
// failure-free cases: a missing key returns an error wrapping ErrNotFound, a
// value that cannot be parsed as T returns the parse error, and only a real
// value comes back with a nil error.
func GetAs[T Parseable](c Config, key string) (T, error) {
	var zero T

	raw, ok := c.Lookup(key)
	if !ok {
		return zero, fmt.Errorf("cfgo: key %q: %w", key, ErrNotFound)
	}

	// A raw value already of type T (e.g. from a Source or Set) passes through.
	if v, ok := raw.(T); ok {
		return v, nil
	}

	s := fmt.Sprintf("%v", raw)
	out := new(T)
	var err error

	switch p := any(out).(type) {
	case *string:
		*p = s
	case *int:
		*p, err = strconv.Atoi(s)
	case *int64:
		*p, err = strconv.ParseInt(s, 10, 64)
	case *float64:
		*p, err = strconv.ParseFloat(s, 64)
	case *bool:
		*p, err = strconv.ParseBool(s)
	case *time.Duration:
		*p, err = time.ParseDuration(s)
	case *[]string:
		*p = splitList(s)
	}

	if err != nil {
		return zero, fmt.Errorf("cfgo: key %q: %w", key, err)
	}
	return *out, nil
}

// GetOr retrieves the value for key from c converted to T, returning def
// when the key is missing or the value cannot be parsed as T.
func GetOr[T Parseable](c Config, key string, def T) T {
	v, err := GetAs[T](c, key)
	if err != nil {
		return def
	}
	return v
}
