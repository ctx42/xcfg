// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package xcfgtest provides test helpers for package xcfg: the [Fake]
// configuration provider and helpers that load a configuration or fail the
// test.
package xcfgtest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

// Fake is a fake configuration provider. Create it with [NewFake] or
// [FakeLoad]; the zero value is not usable.
type Fake struct {
	mx  sync.RWMutex      // Guards map.
	cfg map[string]string // Configuration.
	t   tester.T          // The test manager.
}

var _ xcfg.Provider = &Fake{} // Compile time check.

// NewFake returns a new, empty [Fake].
func NewFake(t tester.T) *Fake {
	return &Fake{
		cfg: make(map[string]string),
		t:   t,
	}
}

// FakeLoad returns a [Fake] holding the configuration [NewFromPath] loads
// from pth. Calls t.Error() and returns nil on error.
func FakeLoad(t tester.T, pth string) *Fake {
	t.Helper()

	cfg := NewFromPath(t, pth)
	if cfg == nil {
		return nil
	}
	return NewFake(t).SetFrom(cfg)
}

// Reader returns a reader over the configuration in the dotenv format, as
// [xcfg.Config.Marshal] writes it.
func (cpf *Fake) Reader() io.Reader {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()

	return strings.NewReader(xcfg.Config(cpf.cfg).Marshal())
}

// Set sets the key to the string form of the value and returns the [Fake].
func (cpf *Fake) Set(key string, val any) *Fake {
	cpf.mx.Lock()
	defer cpf.mx.Unlock()
	cpf.t.Helper()

	cpf.cfg[key] = anyToStr(cpf.t, val)
	return cpf
}

// SetFrom sets configuration from the map.
func (cpf *Fake) SetFrom(m map[string]string) *Fake {
	cpf.mx.Lock()
	defer cpf.mx.Unlock()

	for key, val := range m {
		cpf.cfg[key] = val
	}
	return cpf
}

// SetFromWithPrefix sets configuration from the map. A non-empty prefix is
// added to every key with an underscore separator.
func (cpf *Fake) SetFromWithPrefix(prefix string, m map[string]string) *Fake {
	cpf.mx.Lock()
	defer cpf.mx.Unlock()

	for key, val := range m {
		if prefix != "" {
			key = prefix + "_" + key
		}
		cpf.cfg[key] = val
	}
	return cpf
}

// MustKey returns the value of the key. Calls t.Error() if it is not set.
func (cpf *Fake) MustKey(key string) string {
	cpf.t.Helper()
	val, err := cpf.CfgKey(key)
	if err != nil {
		cpf.t.Error(err)
	}
	return val
}

// DeleteKey removes the key.
func (cpf *Fake) DeleteKey(key string) {
	cpf.mx.Lock()
	defer cpf.mx.Unlock()
	cpf.t.Helper()

	delete(cpf.cfg, key)
}

func (cpf *Fake) CfgKey(key string) (string, error) {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()
	cpf.t.Helper()

	if val, ok := cpf.cfg[key]; ok {
		return val, nil
	}
	return "", xcfg.NewFieldError(key, xcfg.ErrKeyMissing)
}

// Map returns a copy of the configuration map.
func (cpf *Fake) Map() map[string]string {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()

	m := make(map[string]string, len(cpf.cfg))
	for key, val := range cpf.cfg {
		m[key] = val
	}
	return m
}

// Reset removes all configuration keys.
func (cpf *Fake) Reset() {
	cpf.mx.Lock()
	defer cpf.mx.Unlock()
	for k := range cpf.cfg {
		delete(cpf.cfg, k)
	}
}

// Save writes the configuration to the file name in dir and returns its path.
// An empty dir means a temporary directory, removed when the test finishes.
func (cpf *Fake) Save(dir, name string) string {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()
	cpf.t.Helper()

	if dir == "" {
		dir = cpf.t.TempDir()
	}
	pth := filepath.Join(dir, name)

	cfg := xcfg.Config(cpf.cfg)
	data := cfg.Marshal()

	if err := os.WriteFile(pth, []byte(data), 0600); err != nil {
		cpf.t.Error(err.Error())
		return ""
	}
	return pth
}

// Encode encodes the config the same way [xcfg.Config.Encode] does.
func (cpf *Fake) Encode() string {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()

	return xcfg.Config(cpf.cfg).Encode()
}

// SetEnv sets an environment variable for every configuration key using the
// test's Setenv, which restores the previous values when the test ends and
// fails the test on an invalid key.
func (cpf *Fake) SetEnv() {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()
	cpf.t.Helper()

	for k, v := range cpf.cfg {
		cpf.t.Setenv(k, v)
	}
}

// anyToStr returns the string representation of a configuration value.
func anyToStr(t tester.T, val any) string {
	t.Helper()

	var vs string
	switch v := val.(type) {
	case int:
		vs = strconv.Itoa(v)
	case float32:
		vs = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		vs = strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		vs = v
	case time.Time:
		vs = v.Format(time.RFC3339Nano)
	case time.Duration:
		vs = v.String()
	case *time.Location:
		vs = v.String()
	case fmt.Stringer:
		vs = v.String()
	case bool:
		vs = strconv.FormatBool(v)
	default:
		t.Errorf("unsupported config type %#v", val)
	}

	return vs
}
