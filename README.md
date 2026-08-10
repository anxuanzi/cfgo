# CFGO — Environment-Based Configuration for Go

[![CI](https://github.com/anxuanzi/cfgo/actions/workflows/ci.yml/badge.svg)](https://github.com/anxuanzi/cfgo/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/anxuanzi/cfgo.svg)](https://pkg.go.dev/github.com/anxuanzi/cfgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/anxuanzi/cfgo)](https://goreportcard.com/report/github.com/anxuanzi/cfgo)
[![License](https://img.shields.io/github/license/anxuanzi/cfgo)](LICENSE)

CFGO is a lightweight, zero-dependency configuration library for Go. It layers
`.env` files, the process environment, and pluggable sources into one
thread-safe view, with typed access and atomic hot-reload.

## Features

- **Zero dependencies** — standard library only
- **Layered environments** — `.env`, `.{APP_ENV}.env`, personal `.local.env`
- **Thread-safe** — race-tested; `Reload` is atomic (all-or-nothing)
- **Typed access** — classic getters plus generic `GetAs[T]` / `GetOr[T]`
  that surface parse errors instead of silent zeros
- **Configurable loading** — directory (traversal-safe via `os.Root`), file
  set, env selector variable, opt-out of process env
- **Extensible** — any type with `Load() (map[string]any, error)` is a source

## Who's Using CFGO

- [Goe Application Development Framework](https://github.com/oeasenet/goe) — a comprehensive application development framework

## Installation

Requires Go 1.25 or newer.

```bash
go get github.com/anxuanzi/cfgo
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/anxuanzi/cfgo"
)

func main() {
	cfg := cfgo.New()

	dbHost := cfg.GetString("DB_HOST")
	dbPort := cfgo.GetOr(cfg, "DB_PORT", 5432) // default if missing or malformed
	debug := cfg.GetBool("DEBUG_MODE")

	fmt.Printf("Database: %s:%d, Debug: %v\n", dbHost, dbPort, debug)
}
```

Every `Config` method is mirrored as a package-level function operating on a
shared instance (`cfgo.GetString(...)`, `cfgo.Set(...)`, …). The shared
instance is created lazily on the first call — importing cfgo has no side
effects. `cfgo.Default()` returns it if you need to pass it around.

## Precedence

Values are assembled in layers; **later layers win**:

1. `.env` — base configuration
2. `.{APP_ENV}.env` — environment-specific file (`APP_ENV` defaults to `dev`,
   so `.dev.env` is used when it is unset)
3. `.local.env` — personal overrides; keep it out of version control
4. Process environment variables (unless `WithoutSystemEnv()`)
5. Sources added with `AddSource`, in registration order
6. `Set(...)` — programmatic overrides; these also survive `Reload`

Files are read from the current working directory (or the `WithDir`
directory); missing files are skipped. The environment snapshot is taken when
the instance is created — values changed with `os.Setenv` afterwards appear
after the next `Reload`.

### Env-file dialect

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD="quoted values lose their quotes"

# Application
DEBUG_MODE=true
API_TIMEOUT=30s
ALLOWED_ORIGINS=https://example.com,https://api.example.com
```

The dialect is deliberately minimal: lines are `KEY=value`; blank lines and
lines starting with `#` are skipped; keys and values are trimmed, values also
of surrounding single/double quotes. There are **no** inline comments,
`export` prefixes, escape sequences, or multiline values.

## Options

```go
cfg := cfgo.New(
	cfgo.WithDir("configs"),          // load env files from ./configs (os.Root-sandboxed)
	cfgo.WithEnvFiles("app.env"),     // replace the default file set entirely
	cfgo.WithEnvVar("RUN_MODE"),      // use $RUN_MODE instead of $APP_ENV
	cfgo.WithoutSystemEnv(),          // don't import process environment variables
)
```

`Reload` keeps honoring the options the instance was created with.

## Reading values

```go
// Classic getters: zero value when missing OR malformed.
port := cfg.GetInt("PORT")               // 0 if PORT is "808O" — beware
timeout := cfg.GetDuration("API_TIMEOUT")
origins := cfg.GetStringSlice("ALLOWED_ORIGINS") // comma-separated

// Lookup distinguishes absent from empty.
if v, ok := cfg.Lookup("FEATURE_FLAG"); ok {
	fmt.Println("flag present:", v)
}

// Generic getters surface problems instead of hiding them.
port, err := cfgo.GetAs[int](cfg, "PORT") // error on "808O", ErrNotFound if absent
retries := cfgo.GetOr(cfg, "RETRIES", 3)  // default on missing or malformed

// Iterate a consistent snapshot (safe during concurrent writes).
for k, v := range cfg.Each() {
	_ = k
	_ = v
}
```

Validate critical configuration at startup with `GetAs` so typos fail fast
instead of becoming zeros.

### Grouped keys

`GetStringMap("DB")` collects every key under the `DB.` prefix:

```env
# .env — dotted keys work in env files and Set(); most shells cannot
# export dotted names, so don't rely on them via the process environment.
DB.HOST=localhost
DB.PORT=5432
```

```go
db := cfg.GetStringMap("DB") // map[HOST:localhost PORT:5432]
```

## Custom sources

Anything with `Load() (map[string]any, error)` is a source. Implement the
optional `Name() string` (`NamedSource`) to make reload errors identify it.

```go
type JSONSource struct{ Path string }

func (j JSONSource) Name() string { return "json:" + j.Path }

func (j JSONSource) Load() (map[string]any, error) {
	raw, err := os.ReadFile(j.Path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

cfg.AddSource(JSONSource{Path: "config.json"})
if err := cfg.Reload(); err != nil { // AddSource registers; Reload loads
	log.Fatal(err)
}
```

Source values override env files and system environment variables; `Set`
overrides sources. The pre-v1.2 `ConfigSource` interface (with `Name` and
`Watch`) is deprecated but still accepted — cfgo never called `Watch`.

## Reloading

```go
// e.g. after receiving SIGHUP
if err := cfg.Reload(); err != nil {
	log.Printf("config reload failed: %v", err) // previous config still active
}
```

`Reload` is all-or-nothing: it rebuilds from files, environment, and sources,
and swaps the new state in only if everything succeeded. On error (which
names the failing source), the running configuration is untouched. Values set
with `Set` are preserved across reloads.

## Security notes

By default the entire process environment becomes part of the configuration,
so **`All()` and `Each()` will expose secrets** like cloud credentials —
never log them wholesale. Use `WithoutSystemEnv()` to keep the environment
out, and keep secrets in real environment variables or an external source
rather than committed env files.

## Benchmarks

Read-path microbenchmarks on Go 1.26.5, Apple M1 Max (darwin/arm64),
summarized with `benchstat` over 8 runs:

| Operation | Time | Allocations |
|---|---:|---:|
| `Get` | 14.7 ns/op | 0 |
| `Lookup` | 15.4 ns/op | 0 |
| `GetString` | 14.9 ns/op | 0 |
| `GetInt` | 17.3 ns/op | 0 |
| `GetAs[int]` | 52.9 ns/op | 1 (4 B) |
| `Get` under contention¹ | 133 ns/op | 0 |

¹ Deliberate worst case: 10 goroutines reading the same key in a tight loop,
where the `RWMutex` reader count becomes the bottleneck. Occasional reads —
the normal usage pattern — behave like the single-threaded rows. If a value
sits on a genuinely hot per-request path, read it once at startup into your
own struct.

Reproduce with:

```bash
go test -run '^$' -bench . -benchmem -count=8 .
```

## Migrating from v1.1

- `.local.env` now **overrides** `.{APP_ENV}.env` (it previously lost).
- `Set` values now survive `Reload` (they previously vanished).
- The global instance initializes on first use, not at import.
- `AddSource` takes the minimal `Source` interface; existing `ConfigSource`
  implementations compile unchanged. Only third-party implementations of the
  `Config` interface itself need to add `Lookup` and `Each`.

See [CHANGELOG.md](CHANGELOG.md) for the full list, and the
[API reference on pkg.go.dev](https://pkg.go.dev/github.com/anxuanzi/cfgo)
for complete documentation.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Licensed under the Apache License 2.0 — see [LICENSE](LICENSE).
