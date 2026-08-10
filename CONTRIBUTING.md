# Contributing to CFGO

Thanks for helping improve CFGO. A few ground rules keep the library small
and dependable.

## Requirements

- Go 1.25 or newer
- [golangci-lint](https://golangci-lint.run/) v2 (optional locally; CI runs it)

## Development

```bash
go test -race -shuffle=on ./...                 # full suite; hermetic — never touches your real .env files
go test -fuzz FuzzParseEnvLine -fuzztime 30s .  # fuzz the env-file parser
go test -run '^$' -bench . ./...                # benchmarks
golangci-lint run
```

## Guidelines

- **Zero dependencies** is a feature. Changes that add a module dependency
  need a very strong reason.
- **Tests first.** Every behavior change ships with a test that fails
  without it; concurrency-sensitive changes must pass `-race`.
- **Compatibility.** Don't break existing callers; deprecate instead where
  possible, and call out any behavior change in `CHANGELOG.md`.
- Run `gofmt` (CI enforces it via golangci-lint) and keep godoc comments on
  all exported identifiers.

## Pull requests

Explain the *why*, reference issues where relevant, and update the
`Unreleased` section of `CHANGELOG.md` as part of the PR.
