package cfgo

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// loadOnlySource implements only Source (Load) — the minimal contract.
type loadOnlySource struct{ data map[string]any }

func (s loadOnlySource) Load() (map[string]any, error) { return s.data, nil }

// failLoadOnly implements Source, fails, and has no Name.
type failLoadOnly struct{}

func (failLoadOnly) Load() (map[string]any, error) { return nil, errors.New("boom") }

func TestAddSourceAcceptsLoadOnlySource(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	cfg.AddSource(loadOnlySource{data: map[string]any{"MIN_KEY": "min_value"}})
	if err := cfg.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := cfg.GetString("MIN_KEY"); got != "min_value" {
		t.Fatalf("expected 'min_value', got %q", got)
	}
}

func TestReloadErrorFallsBackToSourceTypeName(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	cfg.AddSource(failLoadOnly{})
	err := cfg.Reload()
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "failLoadOnly") {
		t.Fatalf("error should fall back to the source's type name when it has no Name(), got: %v", err)
	}
}

func TestLookup(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	if _, ok := cfg.Lookup("LOOKUP_ABSENT_KEY"); ok {
		t.Fatal("Lookup must report absence with ok=false")
	}
	cfg.Set("LOOKUP_PRESENT_KEY", "v")
	v, ok := cfg.Lookup("LOOKUP_PRESENT_KEY")
	if !ok || v != "v" {
		t.Fatalf("Lookup = (%v, %v), want (v, true)", v, ok)
	}
}

func TestGetAs(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	cfg.Set("PORT", "8080")
	cfg.Set("TIMEOUT", "1h30m")
	cfg.Set("TAGS", "a, b ,c")
	cfg.Set("RAW_INT", 42)
	cfg.Set("BAD_PORT", "808O")

	if v, err := GetAs[int](cfg, "PORT"); err != nil || v != 8080 {
		t.Fatalf("GetAs[int] = (%v, %v)", v, err)
	}
	if v, err := GetAs[time.Duration](cfg, "TIMEOUT"); err != nil || v != 90*time.Minute {
		t.Fatalf("GetAs[time.Duration] = (%v, %v)", v, err)
	}
	if v, err := GetAs[[]string](cfg, "TAGS"); err != nil || len(v) != 3 || v[2] != "c" {
		t.Fatalf("GetAs[[]string] = (%v, %v)", v, err)
	}
	if v, err := GetAs[int](cfg, "RAW_INT"); err != nil || v != 42 {
		t.Fatalf("GetAs[int] with non-string raw value = (%v, %v)", v, err)
	}

	if _, err := GetAs[int](cfg, "BAD_PORT"); err == nil {
		t.Fatal("malformed value must return an error, not a silent zero")
	}
	if _, err := GetAs[int](cfg, "GETAS_NO_SUCH_KEY"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing key must wrap ErrNotFound, got %v", err)
	}
}

func TestGetOr(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := New()
	cfg.Set("PORT", "8080")
	cfg.Set("BAD_PORT", "not-a-number")

	if v := GetOr(cfg, "PORT", 1); v != 8080 {
		t.Fatalf("GetOr with present key = %v, want 8080", v)
	}
	if v := GetOr(cfg, "GETOR_MISSING", 9090); v != 9090 {
		t.Fatalf("GetOr with missing key = %v, want default 9090", v)
	}
	if v := GetOr(cfg, "BAD_PORT", 7070); v != 7070 {
		t.Fatalf("GetOr with malformed value = %v, want default 7070", v)
	}
	if v := GetOr(cfg, "GETOR_MISSING_DUR", 5*time.Second); v != 5*time.Second {
		t.Fatalf("GetOr duration default = %v", v)
	}
}

func TestEach(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, ".env", "EACH_FILE_KEY=file_value")

	cfg := New()
	cfg.Set("EACH_SET_KEY", "set_value")

	got := maps.Collect(cfg.Each())
	if got["EACH_FILE_KEY"] != "file_value" {
		t.Fatalf("Each must yield file-loaded values, got %v", got["EACH_FILE_KEY"])
	}
	if got["EACH_SET_KEY"] != "set_value" {
		t.Fatalf("Each must yield Set overrides, got %v", got["EACH_SET_KEY"])
	}

	// Early break must be safe.
	for range cfg.Each() {
		break
	}
}

func TestWithEnvFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, "base.conf", "LAYERED_KEY=base\nONLY_BASE=1")
	writeEnvFile(t, "extra.conf", "LAYERED_KEY=extra")
	writeEnvFile(t, ".env", "DEFAULT_FILE_KEY=1")

	cfg := New(WithEnvFiles("base.conf", "extra.conf"), WithoutSystemEnv())

	if got := cfg.GetString("LAYERED_KEY"); got != "extra" {
		t.Fatalf("later files must override earlier ones, got %q", got)
	}
	if !cfg.Has("ONLY_BASE") {
		t.Fatal("all listed files must load")
	}
	if cfg.Has("DEFAULT_FILE_KEY") {
		t.Fatal(".env must not load when WithEnvFiles replaces the default set")
	}
}

func TestWithDir(t *testing.T) {
	confDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(confDir, ".env"), []byte("DIR_KEY=dir_value"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir()) // cwd holds no env files

	cfg := New(WithDir(confDir), WithoutSystemEnv())

	if got := cfg.GetString("DIR_KEY"); got != "dir_value" {
		t.Fatalf("WithDir must load env files from the given directory, got %q", got)
	}
}

func TestWithoutSystemEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SYS_ENV_PROBE", "visible")

	if cfg := New(WithoutSystemEnv()); cfg.Has("SYS_ENV_PROBE") {
		t.Fatal("WithoutSystemEnv must exclude process environment variables")
	}
	if cfg := New(); !cfg.Has("SYS_ENV_PROBE") {
		t.Fatal("system env must still be included by default")
	}
}

func TestWithEnvVar(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, ".staging.env", "MODE_KEY=staging_value")
	t.Setenv("RUN_MODE", "staging")
	t.Setenv("APP_ENV", "prod") // must be ignored when WithEnvVar is used

	cfg := New(WithEnvVar("RUN_MODE"), WithoutSystemEnv())

	if got := cfg.GetString("MODE_KEY"); got != "staging_value" {
		t.Fatalf("WithEnvVar must select the environment file via the given variable, got %q", got)
	}
}

func TestReloadRespectsOptions(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SYS_ENV_PROBE_RELOAD", "visible")

	cfg := New(WithoutSystemEnv())
	if err := cfg.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if cfg.Has("SYS_ENV_PROBE_RELOAD") {
		t.Fatal("Reload must keep honoring the options New was built with")
	}
}

// TestDefaultLazyGlobal must be the only test that touches the package-level
// global functions: the first use latches the singleton.
func TestDefaultLazyGlobal(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GLOBAL_LAZY_PROBE", "seen")

	if Default() == nil {
		t.Fatal("Default must return the shared instance")
	}
	if Default() != Default() {
		t.Fatal("Default must return the same instance every time")
	}
	if got := GetString("GLOBAL_LAZY_PROBE"); got != "seen" {
		t.Fatalf("global instance must be created lazily on first use, not at import time; got %q", got)
	}
}
