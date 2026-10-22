// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

func Test_Open(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		pth := "../testdata/example0.conf"

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := Open(tspy, pth)

		// --- Then ---
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("error - not existing file", func(t *testing.T) {
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
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewFromPath(tspy, "../testdata/example0.conf")

		// --- Then ---
		assert.Equal(t, "0", have["VALUE"])
	})

	t.Run("encoded argument", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		enc := must.Value(xcfg.Config{"KEY": "VAL"}.Encode())

		// --- When ---
		have := NewFromPath(tspy, enc)

		// --- Then ---
		assert.Equal(t, xcfg.Config{"KEY": "VAL"}, have)
	})

	t.Run("error - not existing file", func(t *testing.T) {
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

	t.Run("loads after an earlier failure", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("earlier failure")
		tspy.Close()
		tspy.Error("earlier failure")

		// --- When ---
		have := NewFromPath(tspy, "../testdata/example0.conf")

		// --- Then ---
		assert.NotNil(t, have)
	})
}

func Test_NewFromReader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("KEY=VALUE")

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewFromReader(tspy, rdr)

		// --- Then ---
		assert.Equal(t, xcfg.Config{"KEY": "VALUE"}, have)
	})

	t.Run("error - invalid config", func(t *testing.T) {
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
