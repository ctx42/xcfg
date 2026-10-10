// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

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

// Fake is fake configuration provider. Create it with [NewFake] or
// [FakeLoad]; the zero value is not usable.
type Fake struct {
	mx  sync.RWMutex      // Guards map.
	cfg map[string]string // Configuration.
	t   tester.T          // The test manager.
}

var _ xcfg.Provider = &Fake{} // Compile time check.

// NewFake returns new instance of Fake.
func NewFake(t tester.T) *Fake {
	return &Fake{
		cfg: make(map[string]string),
		t:   t,
	}
}

// FakeLoad returns fake based on configuration file.
func FakeLoad(t tester.T, pth string) *Fake {
	t.Helper()

	rdr := Open(t, pth)
	if t.Failed() {
		return nil
	}

	cfg, err := xcfg.New(rdr)
	if err != nil {
		t.Error(err)
		return nil
	}

	cpf := NewFake(t)
	for k, v := range cfg {
		cpf.Set(k, v)
	}
	return cpf
}

func (cpf *Fake) Reader() io.Reader {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()
	cpf.t.Helper()

	vss := make(map[string]string, len(cpf.cfg))
	for k, v := range cpf.cfg {
		vss[k] = anyToStr(cpf.t, v)
	}
	return strings.NewReader(xcfg.Config(vss).Marshal())
}

// Set sets key with value. Implements fluent interface.
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

// SetFromWithPrefix sets configuration from the map. If prefix is provided all
// keys will be prefixed with it and underscore.
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

// MustKey returns configuration key value. Calls t.Error() if it's not set.
func (cpf *Fake) MustKey(key string) string {
	cpf.t.Helper()
	val, err := cpf.CfgKey(key)
	if err != nil {
		cpf.t.Error(err)
	}
	return val
}

// DeleteKey removes keys from internal map.
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
		return anyToStr(cpf.t, val), nil
	}
	return "", xcfg.NewFieldError(key, xcfg.ErrKeyMissing)
}

// Map returns copy of the configuration map.
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

// Save saves the config to the filesystem. If dir is empty string it will
// create temporary directory and put the file with name there. The temporary
// directory will be automatically removed once the test finishes.
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

// SetEnv sets environment variables based on cfg map.
func (cpf *Fake) SetEnv() {
	cpf.mx.RLock()
	defer cpf.mx.RUnlock()
	cpf.t.Helper()

	for k, v := range cpf.cfg {
		if err := os.Setenv(k, v); err != nil {
			cpf.t.Error(err.Error())
		}
	}
}

// anyToStr returns string representation of interface value.
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
