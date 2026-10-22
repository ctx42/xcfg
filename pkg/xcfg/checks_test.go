// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_CheckDirExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "testdata")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - dir does not exist", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "/not/existing/dir")

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KEY_NAME: config key invalid; directory does not exist"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyInvalid, err)
		xrrtest.AssertFieldCode(t, "KEY_NAME", ECCfgKeyInvValue, err)
	})

	t.Run("error - path under a file", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "testdata/equals.conf/sub")

		// --- Then ---
		wMsg := "KEY_NAME: config key invalid; directory does not exist"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyInvalid, err)
		xrrtest.AssertFieldCode(t, "KEY_NAME", ECCfgKeyInvValue, err)
	})

	t.Run("error - dir is a file", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "testdata/equals.conf")

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KEY_NAME: config key invalid; directory does not exist"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyInvalid, err)
		xrrtest.AssertFieldCode(t, "KEY_NAME", ECCfgKeyInvValue, err)
	})

	t.Run("error - empty", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "")

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		assert.ErrorEqual(t, "KEY_NAME: config key not set", err)
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyEmpty, err)
	})
}

func Test_CheckAbsPath(t *testing.T) {
	t.Run("absolute path", func(t *testing.T) {
		// --- When ---
		err := CheckAbsPath("KEY_NAME", "/abs/path")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - relative path", func(t *testing.T) {
		// --- When ---
		err := CheckAbsPath("KEY_NAME", "abs/path")

		// --- Then ---
		wMsg := "KEY_NAME: config key invalid; not an absolute path"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyInvalid, err)
	})

	t.Run("error - empty", func(t *testing.T) {
		// --- When ---
		err := CheckAbsPath("KEY_NAME", "")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyEmpty, err)
		assert.ErrorEqual(t, "KEY_NAME: config key not set", err)
	})
}
