// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package xcfg provides primitives for accessing configuration files.
package xcfg

import (
	"encoding/base64"
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
	// ECCfgError represents general config error code.
	ECCfgError = "ECCfgError"

	// ECCfgKeyInvFormat represents error code for invalid config key format.
	ECCfgKeyInvFormat = "ECCfgKeyInvFormat"

	// ECCfgKeyInvValue represents error code for invalid config key value.
	ECCfgKeyInvValue = "ECCfgKeyInvValue"

	// ECCfgKeyEmpty represents error code for empty config key.
	ECCfgKeyEmpty = "ECCfgKeyEmpty"

	// ECCfgKeyMissing represents error code for missing config key.
	ECCfgKeyMissing = "ECCfgKeyMissing"

	// ECCfgFileNotFound represents error code for config file not found.
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

// ErrCfgFileNotFound is error returned when configuration file was not found
// at given path, or it was not a file or not readable.
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

// Adapter is an interface which adapts configuration keys. See [Adaptor].
type Adapter interface {
	// CfgKeyAdapt adapts configuration key name.
	CfgKeyAdapt(key string) string
}

var _ Provider = Config{} // Compile time check.

// options represent [NewFromPath] options.
type options struct {
	env []string
}

// WithEnv is option for [NewFromPath] setting environment to overwrite keys.
// The keys in the config matching keys present in the environment will be
// overwritten.
func WithEnv(env []string) func(*options) {
	return func(opts *options) { opts.env = env }
}

// Config holds application configuration.
type Config map[string]string

// New creates new instance of Config and loads configuration from reader.
// The reader may hold a Base64 encoded configuration (see [Config.Encode]).
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
func NewFromPath(pth string, opts ...func(*options)) (Config, error) {
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
func (cfg Config) Marshal() string {
	lns := make([]string, 0, len(cfg))
	for k, v := range cfg {
		if d, err := strconv.Atoi(v); err == nil && strconv.Itoa(d) == v {
			lns = append(lns, k+"="+v)
		} else {
			//nolint:gocritic
			lns = append(lns, fmt.Sprintf(`%s="%s"`, k, doubleQuoteEscape(v)))
		}
	}
	sort.Strings(lns)
	return strings.Join(lns, "\n")
}

// Encode encodes the config as an unpadded URL-safe Base64 string with
// the `base64:` prefix.
func (cfg Config) Encode() string {
	data := cfg.Marshal()
	return prefix64 + base64.RawURLEncoding.EncodeToString([]byte(data))
}

// SetFrom sets configuration from the map.
func (cfg Config) SetFrom(m map[string]string) Config {
	for key, val := range m {
		cfg[key] = val
	}
	return cfg
}

// Unmarshal reads a configuration file as a string and returns a key-value map.
func Unmarshal(str string) (Config, error) {
	return New(strings.NewReader(str))
}
