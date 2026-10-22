// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package xcfg provides primitives for accessing configuration files.
package xcfg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/ctx42/dotenv/pkg/dotenv"
	"github.com/ctx42/xrr/pkg/xrr"
)

// Error codes.
const (
	ECCfgError        = "ECCfgError"
	ECCfgKeyInvFormat = "ECCfgKeyInvFormat"
	ECCfgKeyInvValue  = "ECCfgKeyInvValue"
	ECCfgKeyEmpty     = "ECCfgKeyEmpty"
	ECCfgKeyMissing   = "ECCfgKeyMissing"
	ECCfgFileNotFound = "ECCfgFileNotFound"
)

// Sentinel errors.
var (
	ErrKeyMissing   = NewError("config key missing", ECCfgKeyMissing)
	ErrKeyEmpty     = NewError("config key not set", ECCfgKeyEmpty)
	ErrKeyInvalid   = NewError("config key invalid", ECCfgKeyInvValue)
	ErrKeyInvFormat = NewError("config key invalid format", ECCfgKeyInvFormat)
)

// prefix64 is the prefix marking a configuration passed to [New] or
// [NewFromPath] as Base64 encoded (unpadded URL-safe alphabet).
const prefix64 = "base64:"

// ErrCfgFileNotFound is the error returned when the configuration file at the
// given path does not exist, is not a regular file, or is not readable.
var ErrCfgFileNotFound = NewError("config file not found", ECCfgFileNotFound)

// Provider provides configuration values.
type Provider interface {
	// CfgKey returns the value for the given key from configuration without
	// validation. It returns [ErrKeyMissing] if the key does not exist. The
	// key searched in the provider may differ from the provided one if wrapped
	// using adapters like [AddAdapter] or [RemoveAdapter].
	//
	// On error the method must return an instance of [Error] or [FieldErrors].
	CfgKey(key string) (string, error)
}

// Adapter is an interface that adapts configuration keys. See [KeyAdapter].
type Adapter interface {
	// CfgKeyAdapt adapts configuration key name.
	CfgKeyAdapt(key string) string
}

var _ Provider = Config{} // Compile time check.

// Option configures [NewFromPath].
type Option func(*options)

// options holds the settings [Option] values configure.
type options struct {
	env []string
}

// WithEnv is an option for [NewFromPath] that sets the environment used to
// overwrite keys. Configuration keys present in the environment take their
// values from it; other environment variables are ignored.
func WithEnv(env []string) Option {
	return func(opts *options) { opts.env = env }
}

// Config holds application configuration.
type Config map[string]string

// New creates a new [Config] and loads the configuration from the reader.
// Input starting with the `base64:` prefix is always decoded as a Base64
// encoded configuration (see [Config.Encode]), so the prefix is reserved and
// cannot start a plain dotenv file.
func New(r io.Reader) (Config, error) {
	var err error
	cfg := make(map[string]string)
	if r, err = maybeDecode(r); err != nil {
		return nil, NewError("config", ECCfgError, xrr.WithCause(err))
	}
	if err = dotenv.Parse(cfg, r); err != nil {
		return nil, NewError("config", ECCfgError, xrr.WithCause(err))
	}
	return cfg, nil
}

// NewFromPath creates [Config] from `pth`. When `pth` starts with the
// `base64:` prefix, it is treated as the Base64 encoded configuration itself
// rather than a file path.
func NewFromPath(pth string, opts ...Option) (Config, error) {
	def := &options{}
	for _, opt := range opts {
		opt(def)
	}

	var err error
	var cfg Config
	if strings.HasPrefix(pth, prefix64) {
		if cfg, err = New(strings.NewReader(pth)); err != nil {
			return nil, err
		}
	} else {
		var r io.Reader
		if r, err = Open(pth); err != nil {
			err = fmt.Errorf("%w: %w", ErrCfgFileNotFound, err)
			return nil, NewError("", ECCfgFileNotFound, xrr.WithCause(err))
		}
		if cfg, err = New(r); err != nil {
			return nil, err
		}
	}
	// Override matching key values from environment.
	if def.env != nil {
		em := EnvSplit(def.env)
		for ck := range cfg {
			if ev, ok := em[ck]; ok {
				cfg[ck] = ev
			}
		}
	}
	return cfg, nil
}

// NewFromEnv returns configuration based on environment.
func NewFromEnv(env []string) Config { return EnvSplit(env) }

func (cfg Config) CfgKey(key string) (string, error) {
	if v, ok := cfg[key]; ok {
		return v, nil
	}
	return "", NewFieldError(key, ErrKeyMissing)
}

// Marshal outputs the given configuration map as a dotenv-formatted file.
// Each line is in the format: KEY="VALUE" where VALUE is backslash-escaped.
// Keys are written verbatim, so a key [New] cannot read back — empty, or
// holding "=", ":", "#", or whitespace — is an error: a [FieldErrors]
// wrapping [ErrKeyInvalid] for every such key.
func (cfg Config) Marshal() (string, error) {
	var ers map[string]error
	lns := make([]string, 0, len(cfg))
	for k, v := range cfg {
		if !isDotenvKey(k) {
			if ers == nil {
				ers = make(map[string]error)
			}
			ers[k] = keyCause(ErrKeyInvalid, errors.New("not a dotenv key"))
			continue
		}
		if d, err := strconv.Atoi(v); err == nil && strconv.Itoa(d) == v {
			lns = append(lns, k+"="+v)
		} else {
			lns = append(lns, k+`="`+doubleQuoteEscape(v)+`"`)
		}
	}
	if ers != nil {
		return "", NewFieldErrors(ers)
	}
	sort.Strings(lns)
	return strings.Join(lns, "\n"), nil
}

// Encode encodes the config as an unpadded URL-safe Base64 string with
// the `base64:` prefix. Keys follow the same rules as in [Config.Marshal].
func (cfg Config) Encode() (string, error) {
	data, err := cfg.Marshal()
	if err != nil {
		return "", err
	}
	return prefix64 + base64.RawURLEncoding.EncodeToString([]byte(data)), nil
}

// SetFrom sets configuration from the map. On a nil [Config] it allocates a
// new one, so always use the returned value.
func (cfg Config) SetFrom(m map[string]string) Config {
	if cfg == nil {
		cfg = make(Config, len(m))
	}
	for key, val := range m {
		cfg[key] = val
	}
	return cfg
}

// Unmarshal reads a configuration file as a string and returns a key-value map.
func Unmarshal(str string) (Config, error) {
	return New(strings.NewReader(str))
}
