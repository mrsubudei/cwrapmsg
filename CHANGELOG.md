# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2026-11-14

### Changed

- The minimum Go version is 1.23.0, `golang.org/x/tools` is downgraded to v0.36.0.
- `golang.org/x/exp` dependency is removed.

### Added

- CI with tests and linting.

## [1.1.0] - 2026-11-14

### Added

- `-case-insensitive` flag to compare the message with the name case-insensitively.
- `-ignore-cyrillic` flag to skip messages with cyrillic symbols.

### Changed

- The linter is built on `golang.org/x/tools/go/analysis`, run it with `cwrapmsg ./...`.
- The comparison is case-sensitive by default, use `-case-insensitive` for the previous behavior.
- Messages with cyrillic symbols are checked by default, use `-ignore-cyrillic` to skip them.
- Mismatches are reported with a message, the exit code is not zero if mismatches are found.
- Generated files are detected by the `// Code generated ... DO NOT EDIT.` comment too.

### Removed

- `-u`, `-s`, `-c`, `-v` flags.
- `//nolint:cwrapmsg` comments and the `cwrapmsg.txt` exclusion file,
  use the exclusions of golangci-lint or other tools running the analyzer.

## [1.0.0]

- The first release.
