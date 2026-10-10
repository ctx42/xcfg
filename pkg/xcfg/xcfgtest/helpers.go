// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"bytes"
	"io"

	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

// Open returns reader for configuration file. Calls t.Error() on error.
func Open(t tester.T, pth string) *bytes.Reader {
	t.Helper()
	r, err := xcfg.Open(pth)
	if err != nil {
		t.Error(err)
	}
	return r
}

// NewFromPath returns new configuration from path.
// Calls t.Error() and returns nil on error.
func NewFromPath(t tester.T, pth string) xcfg.Config {
	t.Helper()
	rdr := Open(t, pth)
	if rdr == nil {
		return nil
	}

	cfg, err := xcfg.New(rdr)
	if err != nil {
		t.Error(err)
		return nil
	}
	return cfg
}

// NewFromReader returns new configuration form reader.
// Calls t.Error() and returns nil on error.
func NewFromReader(t tester.T, rdr io.Reader) xcfg.Config {
	t.Helper()
	cfg, err := xcfg.New(rdr)
	if err != nil {
		t.Error(err)
		return nil
	}
	return cfg
}
