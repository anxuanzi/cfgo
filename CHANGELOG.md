# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/), and versions follow
[Semantic Versioning](https://semver.org/).

## [1.2.0] - 2026-08-10

### Fixed

- **`Get` no longer serves stale values under concurrent use.** The value
  cache behind `Get` was removed: the released v1.1.0 wrote it under a read
  lock (concurrent reads could crash with `fatal error: concurrent map
  writes`), and the interim fix on `main` used a non-atomic lock upgrade
  that could re-install outdated values after a concurrent `Set` or
  `Reload`. The cache fronted a map lookup with another map lookup under the
  same mutex, so removing it costs nothing. Concurrency regression tests
  added.
- **A failed `Reload` keeps the previous configuration fully intact.** The
  new state is built off-lock and swapped in only on success; previously the
  config was cleared first, so a transient source failure left it partially
  populated.
- `Reload` errors now identify the failing source (via `NamedSource` or the
  source's type) and wrap the underlying error.
- The test suite no longer writes `.env` fixtures into the working copy
  (it could previously overwrite and delete a contributor's real `.env`).
- Env-file scanner errors (e.g. a line exceeding the buffer limit) are
  surfaced through `Reload` instead of being silently ignored.
- README license statement corrected to match the Apache-2.0 `LICENSE` file.

### Changed

- **Precedence: `.local.env` now loads after `.{APP_ENV}.env`**, so personal
  uncommitted overrides beat committed environment files — matching the
  file's documented purpose. Full order: `.env` < `.{APP_ENV}.env` <
  `.local.env` < process environment < sources < `Set`.
- **`Set` values survive `Reload`** and always take precedence over loaded
  values. Previously every `Reload` silently discarded them.
- **The global instance is created lazily on first use** (`sync.OnceValue`)
  instead of in `init()`; importing cfgo no longer reads files or snapshots
  the environment.
- `AddSource` accepts the new minimal `Source` interface. Existing
  `ConfigSource` implementations satisfy it and keep working unchanged.
- The `Config` interface gained `Lookup` and `Each`. This breaks third-party
  types that implement `Config` itself (rare); callers are unaffected.
- `go` directive raised from 1.24 (EOL since 2026-02) to 1.25.0, the oldest
  supported Go release.

### Added

- Functional options for `New`: `WithDir` (rooted via `os.Root`, so lookups
  cannot escape the directory), `WithEnvFiles`, `WithoutSystemEnv`,
  `WithEnvVar`.
- `Lookup(key) (any, bool)` — distinguishes absent from present-but-empty.
- Generic getters `GetAs[T]` / `GetOr[T]` with `ErrNotFound` sentinel:
  malformed values return errors (or the default) instead of silent zeros.
- `Each() iter.Seq2[string, any]` — snapshot iteration without copying into
  user code, safe against concurrent mutation.
- `Default()` — access to the shared instance used by package-level
  functions.
- `Source` (minimal, `Load`-only) and optional `NamedSource` interfaces;
  `ConfigSource` is deprecated (its `Watch` was never called by cfgo).
- CI (GitHub Actions: race tests on the two supported Go branches across
  three OSes, golangci-lint v2, govulncheck), fuzz target for the env-file
  parser, runnable pkg.go.dev examples, benchmarks.
- Package documentation (`doc.go`), `CONTRIBUTING.md`, `SECURITY.md`, this
  changelog.

## [1.1.0] - 2025-06-11

### Added

- Global configuration instance with package-level helper functions.

### Known issues (fixed in Unreleased)

- `Get` writes a cache under a read lock; concurrent reads can crash with
  `fatal error: concurrent map writes`. Do not use this version in
  concurrent programs.

## [1.0.2] - 2024-05-16

- `GetArray` function added to the godotenv-based implementation.

## [1.0.1] - 2024-05-16

- `GetArray` method added to the Config interface.

## [1.0.0] - 2024-05-16

- Initial release: godotenv-based loading with a singleton Config instance.
