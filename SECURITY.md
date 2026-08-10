# Security Policy

## Supported versions

The latest release receives security fixes. Older minor versions are not
maintained. Note that releases before the post-1.1.0 concurrency fix can
crash under concurrent reads — upgrade.

## Reporting a vulnerability

Please use GitHub's private vulnerability reporting on this repository
(Security tab → "Report a vulnerability") instead of opening a public issue.
Reports are typically acknowledged within a week.

## Notes for users

- By default the entire process environment is imported into the
  configuration, so `All()` and `Each()` expose every environment variable —
  including secrets. Never log their output wholesale; use
  `WithoutSystemEnv()` where that risk matters.
- `WithDir` opens the directory via `os.Root`, so env-file lookups cannot
  escape it through symlinks or path traversal.
- CI runs `govulncheck` on every push; the module itself has no
  dependencies.
