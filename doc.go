// Package cfgo is a lightweight, zero-dependency configuration library for
// Go applications, built around env files, the process environment, and
// pluggable sources.
//
// # Loading and precedence
//
// A Config assembles values from several layers; when the same key appears
// in more than one, the later layer wins:
//
//	.env                 base configuration
//	.{APP_ENV}.env       environment-specific file (APP_ENV defaults to "dev")
//	.local.env           personal, uncommitted overrides
//	process environment  unless WithoutSystemEnv is used
//	sources              registered with AddSource, in registration order
//	Set                  programmatic overrides; survive Reload
//
// Env files are read from the current working directory unless WithDir
// points elsewhere. Missing files are skipped silently.
//
// # The env-file dialect
//
// Files are parsed line by line: blank lines and lines starting with # are
// skipped; the first '=' splits key from value; both are trimmed of
// whitespace, and the value additionally of surrounding single or double
// quotes. There are no inline comments, no "export" prefixes, no escape
// sequences, and no multiline values.
//
// # Reading values
//
// The classic getters (GetInt, GetBool, ...) return zero values for missing
// and malformed entries alike. When the difference matters, use Lookup, or
// the generic GetAs and GetOr, which report parse errors and wrap
// ErrNotFound for missing keys.
//
// # Concurrency and reloading
//
// All methods are safe for concurrent use. Reload is atomic: readers see
// either the previous or the new state, never a partial one, and a failed
// reload leaves the previous configuration intact.
//
// # The global instance
//
// Package-level functions mirror the Config API on a shared instance
// created lazily on first use, so importing cfgo has no side effects. The
// instance snapshots the environment when created; changes made with
// os.Setenv afterwards become visible only after Reload.
//
// # Security
//
// By default the process environment is part of the configuration, so All
// and Each expose every environment variable, including secrets. Avoid
// logging them wholesale, or construct instances with WithoutSystemEnv.
package cfgo
