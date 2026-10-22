# AGENTS.md

This file provides guidance to coding agents when working with code in this
repository.

## Overview

Module `github.com/ctx42/xcfg` is a Go library for key-value application
configuration. There is no executable. It has two packages:

- `pkg/xcfg` — loading, key adapters, value parsers, and errors.
- `pkg/xcfg/xcfgtest` — test helpers: the `Fake` provider and the
  load-or-fail helpers `Open`, `NewFromPath`, and `NewFromReader`.

Dependencies are limited to ctx42 modules (`dotenv`, `xrr`, and, in tests,
`testing` and `testkit`). Do not add third-party modules.

## Commands

```bash
go test -count=1 -race ./...                  # run all tests
go test -count=1 -run Test_New ./pkg/xcfg/    # run a single top-level test
go test -count=1 -run 'Test_New/parse_plain' ./pkg/xcfg/  # run a subtest
go vet ./...                                  # vet
golangci-lint run                             # lint (config is CI-provided)
```

## Architecture

- **`xcfg.go`** — the `Provider` interface (`CfgKey`), the `Adapter`
  interface (`CfgKeyAdapt`), and `Config` (`map[string]string`, the base
  provider). `New` parses dotenv text through `dotenv.Parse`.
  `NewFromPath` reads a file, or treats a `base64:` argument as the encoded
  configuration itself, then applies `WithEnv` overrides to keys already
  present. Also holds the error codes and sentinel errors.
- **`key_adapter.go`** — `KeyAdapter` wraps a `Provider` and rewrites the key
  before the lookup. `CfgKey` applies only its own rewrite and delegates, so
  each nested adapter rewrites the key once; `CfgKeyAdapt` returns the key
  after every nested rewrite.
- **`parsers.go`** — `Parse*` turn a raw value into a typed one; every error
  is a `*FieldErrors` keyed by the configuration key.
- **`error_types.go`** — the package's `xrr` error domain: `Error`,
  `FieldErrors`, their constructors, and `IsConfigError`.
- **`keys.go`** — key-prefix functions and `GetOr`.
- **`checks.go`** — path validators `CheckDirExists` and `CheckAbsPath`.
- **`helpers.go`** — generic helpers (`Open`, `Merge`, `DirExists`,
  `EnvSplit`) and the unexported `maybeDecode` and `doubleQuoteEscape`.

### Encoded configuration

`Config.Encode` output is `base64:` followed by `Config.Marshal` output in
`base64.RawURLEncoding` (unpadded, URL-safe). `maybeDecode` decodes the same
format. Changing either side breaks configurations already encoded by users.

### Errors

Every `Provider.CfgKey` implementation must return an `Error` or a
`*FieldErrors` on failure; a missing key is
`NewFieldError(key, ErrKeyMissing)`.

## Conventions

- Tests use `github.com/ctx42/testing` (`assert`, `must`, `tester`) and
  `github.com/ctx42/xrr/pkg/xrr/xrrtest`, with
  `// --- Given ---` / `// --- When ---` / `// --- Then ---` blocks.
- Test helpers take `tester.T` and are tested with `tester.Spy`.
- Unexported functions are tested from the white-box `package xcfg` tests;
  `xcfg_test.go` and `example_test.go` are black-box.
- Runnable examples live in `pkg/xcfg/example_test.go`; the README embeds
  them through `gmmce` markers, so keep both in sync.
- Line width is 80 columns (`.editorconfig`).
