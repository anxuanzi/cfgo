package cfgo_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/anxuanzi/cfgo"
)

func Example() {
	// A fresh instance; the package-level functions mirror this API on a
	// shared lazy instance.
	cfg := cfgo.New()
	cfg.Set("CFGO_EXAMPLE_APP_NAME", "orders-api")

	fmt.Println(cfg.GetString("CFGO_EXAMPLE_APP_NAME"))
	// Output: orders-api
}

func ExampleGetOr() {
	cfg := cfgo.New()
	cfg.Set("CFGO_EXAMPLE_PORT", "8080")
	cfg.Set("CFGO_EXAMPLE_BAD_TIMEOUT", "not-a-duration")

	// Present and parseable: the parsed value.
	fmt.Println(cfgo.GetOr(cfg, "CFGO_EXAMPLE_PORT", 3000))
	// Missing: the default.
	fmt.Println(cfgo.GetOr(cfg, "CFGO_EXAMPLE_MISSING", 3000))
	// Malformed: the default, instead of a silent zero.
	fmt.Println(cfgo.GetOr(cfg, "CFGO_EXAMPLE_BAD_TIMEOUT", 5*time.Second))
	// Output:
	// 8080
	// 3000
	// 5s
}

func ExampleGetAs() {
	cfg := cfgo.New()
	cfg.Set("CFGO_EXAMPLE_TIMEOUT", "1m30s")

	timeout, err := cfgo.GetAs[time.Duration](cfg, "CFGO_EXAMPLE_TIMEOUT")
	fmt.Println(timeout, err)

	_, err = cfgo.GetAs[int](cfg, "CFGO_EXAMPLE_ABSENT")
	fmt.Println(errors.Is(err, cfgo.ErrNotFound))
	// Output:
	// 1m30s <nil>
	// true
}

func ExampleConfig_Lookup() {
	cfg := cfgo.New()
	cfg.Set("CFGO_EXAMPLE_EMPTY", "")

	if v, ok := cfg.Lookup("CFGO_EXAMPLE_EMPTY"); ok {
		fmt.Printf("present: %q\n", v)
	}
	if _, ok := cfg.Lookup("CFGO_EXAMPLE_ABSENT"); !ok {
		fmt.Println("absent")
	}
	// Output:
	// present: ""
	// absent
}

func ExampleConfig_AddSource() {
	cfg := cfgo.New()

	// Any type with Load() (map[string]any, error) is a Source.
	cfg.AddSource(staticSource{"CFGO_EXAMPLE_FEATURE": "on"})
	if err := cfg.Reload(); err != nil {
		fmt.Println("reload failed:", err)
		return
	}

	fmt.Println(cfg.GetString("CFGO_EXAMPLE_FEATURE"))
	// Output: on
}

type staticSource map[string]any

func (s staticSource) Load() (map[string]any, error) { return s, nil }
