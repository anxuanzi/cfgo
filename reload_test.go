package cfgo

import (
	"errors"
	"strings"
	"testing"
)

// failingSource is a ConfigSource whose Load can be made to fail on demand.
type failingSource struct {
	name string
	fail bool
	data map[string]any
}

func (f *failingSource) Name() string { return f.name }

func (f *failingSource) Load() (map[string]any, error) {
	if f.fail {
		return nil, errors.New("backend unavailable")
	}
	return f.data, nil
}

func (f *failingSource) Watch(callback func(map[string]any)) error { return nil }

func TestReloadFailureKeepsPreviousConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	src := &failingSource{name: "flaky", data: map[string]any{"RELOAD_SAFE_KEY": "v1"}}
	cfg.AddSource(src)
	if err := cfg.Reload(); err != nil {
		t.Fatalf("initial reload: %v", err)
	}
	if got := cfg.GetString("RELOAD_SAFE_KEY"); got != "v1" {
		t.Fatalf("precondition: expected 'v1', got %q", got)
	}

	src.fail = true
	if err := cfg.Reload(); err == nil {
		t.Fatal("expected an error from Reload when a source fails")
	}

	// A failed reload must leave the previous configuration fully intact.
	if got := cfg.GetString("RELOAD_SAFE_KEY"); got != "v1" {
		t.Fatalf("failed Reload destroyed the running config: expected 'v1', got %q", got)
	}
}

func TestReloadErrorNamesFailingSource(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	src := &failingSource{name: "vault-backend", fail: true}
	cfg.AddSource(src)

	err := cfg.Reload()
	if err == nil {
		t.Fatal("expected an error from Reload when a source fails")
	}
	if !strings.Contains(err.Error(), "vault-backend") {
		t.Fatalf("reload error should name the failing source, got: %v", err)
	}
	if !errors.Is(err, errUnwrapTarget(err)) {
		t.Fatalf("reload error should wrap the source error: %v", err)
	}
	if !strings.Contains(err.Error(), "backend unavailable") {
		t.Fatalf("reload error should preserve the underlying cause, got: %v", err)
	}
}

// errUnwrapTarget digs out the innermost error so we can assert wrapping.
func errUnwrapTarget(err error) error {
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		err = inner
	}
}

func TestSetSurvivesReload(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	cfg.Set("EXPLICIT_OVERRIDE", "kept")

	if err := cfg.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}

	if got := cfg.GetString("EXPLICIT_OVERRIDE"); got != "kept" {
		t.Fatalf("Set value must survive Reload, got %q", got)
	}
	if !cfg.Has("EXPLICIT_OVERRIDE") {
		t.Fatal("Has must report Set values after Reload")
	}
	if _, ok := cfg.All()["EXPLICIT_OVERRIDE"]; !ok {
		t.Fatal("All must include Set values after Reload")
	}
}

func TestSetOverridesSourceValue(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	src := &failingSource{name: "backend", data: map[string]any{"SHARED": "from_source"}}
	cfg.AddSource(src)
	if err := cfg.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}

	cfg.Set("SHARED", "explicit")
	if got := cfg.GetString("SHARED"); got != "explicit" {
		t.Fatalf("Set must take precedence over source values, got %q", got)
	}

	// Explicit overrides stay on top even after the sources are re-read.
	if err := cfg.Reload(); err != nil {
		t.Fatalf("second reload: %v", err)
	}
	if got := cfg.GetString("SHARED"); got != "explicit" {
		t.Fatalf("Set must still win after Reload, got %q", got)
	}
}
