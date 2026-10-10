// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
)

func Test_Open(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		pth := "../testdata/example0.conf"

		// --- When ---
		rdr := Open(t, pth)

		// --- Then ---
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, rdr))
	})

	t.Run("fatal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		have := Open(tspy, "not_existing")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_NewFromPath(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- When ---
		cfg := NewFromPath(t, "../testdata/example0.conf")
		val, err := cfg.CfgKey("VALUE")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "0", val)
	})

	t.Run("fatal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		have := NewFromPath(tspy, "not_existing")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_NewFromReader(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("KEY=VALUE")

		// --- When ---
		cfg := NewFromReader(t, rdr)
		val, err := cfg.CfgKey("KEY")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "VALUE", val)
	})

	t.Run("fatal", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("---")

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("config: dotenv: invalid config line: 1")
		tspy.Close()

		// --- When ---
		have := NewFromReader(tspy, rdr)

		// --- Then ---
		assert.Nil(t, have)
	})
}
