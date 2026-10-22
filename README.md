[![Go Reference](https://pkg.go.dev/badge/github.com/ctx42/xcfg.svg)](https://pkg.go.dev/github.com/ctx42/xcfg)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ctx42/xcfg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

# xcfg

Key-value configuration from dotenv files, environment, and encoded strings.

xcfg loads application configuration in the dotenv format into a plain
`map[string]string` and reads it through a one-method `Provider` interface.
Key adapters remap key names so a component can read its own unprefixed keys
from a shared configuration, and parsers turn raw values into typed ones
with errors that name the offending key.

## Features

- `Provider`: a one-method interface, `CfgKey(key) (string, error)`.
- Loading from a reader (`New`), a file (`NewFromPath`), or an environment
  list (`NewFromEnv`).
- Environment overrides: `WithEnv` overwrites file keys with matching
  environment values.
- Encoded configurations: `Config.Encode` and `base64:` input carry a whole
  configuration in one string.
- Key adapters that add, strip, or change prefixes, or translate keys
  through a map.
- Key helpers: prefix functions for single keys and whole maps
  (`StripPrefixes`, `AddPrefixes`, `ChangePrefix`), `GetOr` defaults, and
  `Merge`.
- Typed parsers: `ParseInteger`, `ParseBool`, `ParseDur`, `ParseTime`,
  `ParseTimezone`.
- Validators: `CheckAbsPath`, `CheckDirExists`.
- Structured errors: error codes and per-key field errors built on
  `ctx42/xrr`.
- Test fakes: `xcfgtest.Fake`, a mutable, goroutine-safe `Provider`.

## Prerequisites

Go 1.26 or later.

## Install

```bash
go get github.com/ctx42/xcfg
```

```go
import "github.com/ctx42/xcfg/pkg/xcfg"
```

## Usage

### Loading and reading keys

<!-- gmmce:pkg/xcfg/ExampleNew -->
```go
src := "# Database.\nDB_HOST=localhost\nDB_PORT=5432\n"
cfg, _ := xcfg.New(strings.NewReader(src))

host, _ := cfg.CfgKey("DB_HOST")
port, _ := xcfg.ParseInteger("DB_PORT", cfg["DB_PORT"])
fmt.Println(host, port)

_, err := cfg.CfgKey("DB_USER")
fmt.Println(err)
// Output:
// localhost 5432
// DB_USER: config key missing
```

The file syntax — quoting, `export` prefixes, `${VAR}` expansion — is that of
[ctx42/dotenv](https://github.com/ctx42/dotenv).

### Encoded configuration and environment overrides

`Config.Encode` packs a whole configuration into a single string: the
`base64:` prefix followed by the dotenv text in unpadded URL-safe Base64.
`New` decodes such input transparently, and `NewFromPath` accepts it in place
of a file path, which is convenient for passing a configuration through one
environment variable or flag. `WithEnv` then overwrites keys that also appear
in the given environment.

<!-- gmmce:pkg/xcfg/ExampleNewFromPath -->
```go
enc := xcfg.Config{"LOG_LEVEL": "info"}.Encode()
env := []string{"LOG_LEVEL=debug"}

cfg, _ := xcfg.NewFromPath(enc, xcfg.WithEnv(env))

fmt.Println(enc)
fmt.Println(cfg["LOG_LEVEL"])
// Output:
// base64:TE9HX0xFVkVMPSJpbmZvIg
// debug
```

To encode an existing file from the shell:

```bash
echo "base64:$(basenc --base64url -w0 app.conf | tr -d '=')"
```

### Key adapters

Adapters wrap a `Provider` and rewrite the key before the lookup. They chain,
and a missing-key error names the key as finally looked up.

<!-- gmmce:pkg/xcfg/ExampleAddAdapter -->
```go
cfg := xcfg.Config{"DB_HOST": "localhost"}
db := xcfg.AddAdapter(cfg, "DB")

host, _ := db.CfgKey("HOST")
fmt.Println(host)
// Output:
// localhost
```

| Adapter         | Key asked for | Key looked up                       |
|-----------------|---------------|-------------------------------------|
| `AddAdapter`    | `HOST`        | `DB_HOST`                           |
| `RemoveAdapter` | `DB_HOST`     | `HOST`                              |
| `ChangeAdapter` | `OLD_HOST`    | `NEW_HOST` (others unchanged)       |
| `MapAdapter`    | `HOST`        | the mapped key, if `HOST` is in map |

### Errors

Errors from loading, providers, parsers, and checks belong to the package's
`xrr` error domain (`IsConfigError`) and carry a code: `ECCfgError`,
`ECCfgKeyMissing`, `ECCfgKeyEmpty`, `ECCfgKeyInvFormat`, `ECCfgKeyInvValue`,
or `ECCfgFileNotFound`. Key-level errors are `*FieldErrors` keyed by the
configuration key and wrap one of the sentinels `ErrKeyMissing`,
`ErrKeyEmpty`, `ErrKeyInvalid`, or `ErrKeyInvFormat`, so
`errors.Is(err, xcfg.ErrKeyInvFormat)` works through them.

### Testing

Package `xcfgtest` provides `Fake`, a goroutine-safe `Provider` whose keys a
test sets, deletes, saves to a file, or exports to the environment, plus
helpers that load a configuration or fail the test.

```go
import "github.com/ctx42/xcfg/pkg/xcfg/xcfgtest"
```

Build one with `xcfgtest.NewFake(t).Set("DB_PORT", 5432)`. `Set` takes a
string, `int`, float, `bool`, `time.Time`, `time.Duration`, `*time.Location`,
or any `fmt.Stringer`.
