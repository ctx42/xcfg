// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

func Test_NewFake(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t, 0)
	tspy.Close()

	// --- When ---
	have := NewFake(tspy)

	// --- Then ---
	assert.Equal(t, map[string]string{}, have.Map())
}

func Test_FakeLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := FakeLoad(tspy, "../testdata/example0.conf")

		// --- Then ---
		want := map[string]string{
			"ABS_DIR": "/some/absolute/path/log",
			"LOG_DIR": "tmp/log",
			"VALUE":   "0",
		}
		assert.Equal(t, want, have.Map())
	})

	t.Run("error - not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		have := FakeLoad(tspy, "not_existing_file")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("error - invalid config", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("dotenv: invalid config line: 1")
		tspy.Close()

		// --- When ---
		have := FakeLoad(tspy, "testdata/invalid.conf")

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
		have := FakeLoad(tspy, "../testdata/example0.conf")

		// --- Then ---
		assert.NotNil(t, have)
	})
}

func Test_Fake_Reader(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := FakeLoad(tspy, "../testdata/example0.conf")

	// --- When ---
	have := cpf.Reader()

	// --- Then ---
	want := "" +
		"ABS_DIR=\"/some/absolute/path/log\"\n" +
		"LOG_DIR=\"tmp/log\"\n" +
		"VALUE=0"
	assert.Equal(t, want, iokit.ReadAllStr(t, have))
}

func Test_Fake_Set(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		tt := []struct {
			testN string

			val  any
			want string
		}{
			{"string", "abc", "abc"},
			{"int", 1, "1"},
			{"float32", float32(1.23), "1.23"},
			{"float64", 1.23, "1.23"},
			{"duration", time.Second, "1s"},
			{"location", time.UTC, "UTC"},
			{
				"time",
				time.Date(2021, 9, 10, 13, 23, 24, 123, time.UTC),
				"2021-09-10T13:23:24.000000123Z",
			},
		}

		for _, tc := range tt {
			t.Run(tc.testN, func(t *testing.T) {
				// --- Given ---
				tspy := tester.New(t)
				tspy.Close()

				cpf := NewFake(tspy)

				// --- When ---
				cpf.Set("KEY", tc.val)

				// --- Then ---
				assert.Equal(t, tc.want, must.Value(cpf.CfgKey("KEY")))
			})
		}
	})

	t.Run("error - unsupported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("unsupported config type %#v", uint(10))
		tspy.Close()

		cpf := NewFake(tspy)

		// --- When ---
		cpf.Set("KEY", uint(10))

		// --- Then ---
		val, err := cpf.CfgKey("KEY")
		xrrtest.AssertFieldIs(t, "KEY", xcfg.ErrKeyMissing, err)
		assert.Equal(t, "", val)
	})
}

func Test_Fake_SetFrom(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := NewFake(tspy)
	cpf.Set("KEY0", "VAL0")

	m := map[string]string{
		"KEY0": "VAL99",
		"KEY1": "VAL1",
	}

	// --- When ---
	cpf.SetFrom(m)

	// --- Then ---
	assert.Equal(t, "VAL99", must.Value(cpf.CfgKey("KEY0")))

	assert.Equal(t, "VAL1", must.Value(cpf.CfgKey("KEY1")))
}

func Test_Fake_SetFromWithPrefix(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("KEY0", "VAL0")

		m := map[string]string{
			"KEY0": "VAL99",
			"KEY1": "VAL1",
		}

		// --- When ---
		cpf.SetFromWithPrefix("", m)

		// --- Then ---
		assert.Equal(t, "VAL99", must.Value(cpf.CfgKey("KEY0")))

		assert.Equal(t, "VAL1", must.Value(cpf.CfgKey("KEY1")))
	})

	t.Run("with prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("KEY0", "VAL0")

		m := map[string]string{
			"KEY0": "VAL99",
			"KEY1": "VAL1",
		}

		// --- When ---
		cpf.SetFromWithPrefix("PREF", m)

		// --- Then ---
		assert.Equal(t, "VAL99", must.Value(cpf.CfgKey("PREF_KEY0")))

		assert.Equal(t, "VAL1", must.Value(cpf.CfgKey("PREF_KEY1")))
	})

	t.Run("empty key with prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		cpf := NewFake(tspy)

		// --- When ---
		cpf.SetFromWithPrefix("PREF", map[string]string{"": "VAL"})

		// --- Then ---
		assert.Equal(t, map[string]string{"PREF": "VAL"}, cpf.Map())
	})
}

func Test_Fake_MustKey(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := FakeLoad(tspy, "../testdata/example0.conf")

		// --- When ---
		have := cpf.MustKey("VALUE")

		// --- Then ---
		assert.Equal(t, "0", have)
	})

	t.Run("error - missing key", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("NOT_EXISTING: config key missing")
		tspy.Close()

		cpf := FakeLoad(tspy, "../testdata/example0.conf")

		// --- When ---
		have := cpf.MustKey("NOT_EXISTING")

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_Fake_DeleteKey(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := FakeLoad(tspy, "../testdata/example0.conf")

	// --- When ---
	cpf.DeleteKey("LOG_DIR")

	// --- Then ---
	val, err := cpf.CfgKey("LOG_DIR")
	xrrtest.AssertFieldIs(t, "LOG_DIR", xcfg.ErrKeyMissing, err)
	assert.Equal(t, "", val)
}

func Test_Fake_CfgKey(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy).Set("KEY", "VAL")

		// --- When ---
		have, err := cpf.CfgKey("KEY")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "VAL", have)
	})

	t.Run("error - missing key", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)

		// --- When ---
		have, err := cpf.CfgKey("KEY")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "KEY", xcfg.ErrKeyMissing, err)
		assert.Equal(t, "", have)
	})
}

func Test_Fake_Map(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("KEY0", "VAL0")
		cpf.Set("KEY1", "VAL1")

		// --- When ---
		have := cpf.Map()

		// --- Then ---
		want := map[string]string{
			"KEY0": "VAL0",
			"KEY1": "VAL1",
		}
		assert.Equal(t, want, have)
	})

	t.Run("empty config", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		cpf := NewFake(tspy)

		// --- When ---
		have := cpf.Map()

		// --- Then ---
		want := map[string]string{}
		assert.Equal(t, want, have)
	})
}

func Test_Fake_Reset(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := FakeLoad(tspy, "../testdata/example0.conf")

	// --- When ---
	cpf.Reset()

	// --- Then ---
	val, err := cpf.CfgKey("LOG_DIR")
	xrrtest.AssertFieldIs(t, "LOG_DIR", xcfg.ErrKeyMissing, err)
	assert.Equal(t, "", val)

	val, err = cpf.CfgKey("ABS_DIR")
	xrrtest.AssertFieldIs(t, "ABS_DIR", xcfg.ErrKeyMissing, err)
	assert.Equal(t, "", val)

	val, err = cpf.CfgKey("VALUE")
	xrrtest.AssertFieldIs(t, "VALUE", xcfg.ErrKeyMissing, err)
	assert.Equal(t, "", val)
}

func Test_Fake_Save(t *testing.T) {
	t.Run("empty dir argument", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectTempDir(1)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("LOG_DIR", "/var/log")
		cpf.Set("VALUE", "value")

		// --- When ---
		pth := cpf.Save("", "my.conf")

		// --- Then ---
		want := "LOG_DIR=\"/var/log\"\nVALUE=\"value\""
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		tspy.Finish()
		assert.Equal(t, filepath.Join(tspy.GetTempDir(0), "my.conf"), pth)
	})

	t.Run("dir argument provided", func(t *testing.T) {
		// --- Given ---
		dir := t.TempDir()

		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("LOG_DIR", "/var/log")
		cpf.Set("VALUE", "value")

		// --- When ---
		pth := cpf.Save(dir, "my.conf")

		// --- Then ---
		assert.Equal(t, filepath.Join(dir, "my.conf"), pth)
		want := "" +
			"LOG_DIR=\"/var/log\"\n" +
			"VALUE=\"value\""
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))
	})

	t.Run("error - path is a directory", func(t *testing.T) {
		// --- Given ---
		dir := t.TempDir()
		cfgPth := oskit.MkdirAll(t, dir, "dir")

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("is a directory")
		tspy.ExpectLogContain(cfgPth)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("LOG_DIR", "/var/log")
		cpf.Set("VALUE", "value")

		// --- When ---
		pth := cpf.Save(dir, "dir")

		// --- Then ---
		assert.Empty(t, pth)
	})
}

func Test_Fake_Encode(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := NewFake(tspy)
	cpf.Set("LOG_DIR", "/var/log")
	cpf.Set("VALUE", "value")

	// --- When ---
	have := cpf.Encode()

	// --- Then ---
	want := "base64:TE9HX0RJUj0iL3Zhci9sb2ciClZBTFVFPSJ2YWx1ZSI"
	assert.Equal(t, want, have)

	check := must.Value(base64.RawURLEncoding.DecodeString(want[7:]))
	want = "" +
		"LOG_DIR=\"/var/log\"\n" +
		"VALUE=\"value\""
	assert.Equal(t, want, string(check))
}

func Test_Fake_SetEnv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectSetenv("SUB_TEST_LOG_DIR", "/var/log")
		tspy.ExpectSetenv("SUB_TEST_VALUE", "value")
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("SUB_TEST_LOG_DIR", "/var/log")
		cpf.Set("SUB_TEST_VALUE", "value")

		// --- When ---
		cpf.SetEnv()

		// --- Then ---
		val, ok := os.LookupEnv("SUB_TEST_LOG_DIR")
		assert.True(t, ok)
		assert.Equal(t, "/var/log", val)

		val, ok = os.LookupEnv("SUB_TEST_VALUE")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("restored after test", func(t *testing.T) {
		// --- When ---
		t.Run("set", func(t *testing.T) {
			NewFake(t).Set("SUB_TEST_RESTORED", "value").SetEnv()
		})

		// --- Then ---
		_, ok := os.LookupEnv("SUB_TEST_RESTORED")
		assert.False(t, ok)
	})
}

func Test_anyToStr_tabular(t *testing.T) {

	tt := []struct {
		testN string

		in   any
		want string
	}{
		{"int", 123, "123"},
		{"float32", float32(123), "123"},
		{"float64", float64(123), "123"},
		{"string", "abc", "abc"},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"2000-01-02T03:04:05Z",
		},
		{"location", time.UTC, "UTC"},
		{"stringer", TStringer(123), "123"},
		{"bool", true, "true"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			// --- When ---
			have, ok := anyToStr(tspy, tc.in)

			// --- Then ---
			assert.True(t, ok)
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_anyToStr(t *testing.T) {
	t.Run("error - unsupported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("unsupported config type xcfgtest.Abc{}")
		tspy.Close()

		type Abc struct{}

		// --- When ---
		have, ok := anyToStr(tspy, Abc{})

		// --- Then ---
		assert.False(t, ok)
		assert.Empty(t, have)
	})
}
