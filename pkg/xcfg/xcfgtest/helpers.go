// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"io"

	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

// Open returns reader for configuration file. Calls t.Error() on error.
func Open(t tester.T, pth string) io.Reader {
	t.Helper()
	r, err := xcfg.Open(pth)
	if err != nil {
		t.Error(err)
	}
	return r
}

// NewFromPath returns new configuration provider from path.
// Calls t.Error() on error.
func NewFromPath(t tester.T, pth string) xcfg.Provider {
	t.Helper()
	rdr := Open(t, pth)
	if t.Failed() {
		return nil
	}

	cfg, err := xcfg.New(rdr)
	if err != nil {
		t.Error(err)
	}
	return cfg
}

// NewFromReader returns new configuration provider form reader.
// Calls t.Error() on error.
func NewFromReader(t tester.T, rdr io.Reader) xcfg.Provider {
	t.Helper()
	cfg, err := xcfg.New(rdr)
	if err != nil {
		t.Error(err)
	}
	return cfg
}
