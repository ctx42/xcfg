// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"bytes"
	"io"

	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

// Open returns a reader for the configuration file. Calls t.Error() on error.
func Open(t tester.T, pth string) *bytes.Reader {
	t.Helper()
	r, err := xcfg.Open(pth)
	if err != nil {
		t.Error(err)
	}
	return r
}

// NewFromPath returns the configuration loaded the way [xcfg.NewFromPath]
// loads it, so pth may also be a `base64:` encoded configuration. Calls
// t.Error() and returns nil on error.
func NewFromPath(t tester.T, pth string) xcfg.Config {
	t.Helper()
	cfg, err := xcfg.NewFromPath(pth)
	if err != nil {
		t.Error(err)
		return nil
	}
	return cfg
}

// NewFromReader returns the configuration loaded from the reader.
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
