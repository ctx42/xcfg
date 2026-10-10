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

func Test_FakeLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		cpf := FakeLoad(t, "../testdata/example0.conf")

		// --- Then ---
		val, err := cpf.CfgKey("VALUE")
		assert.Equal(t, "0", val)
		assert.NoError(t, err)

		val, err = cpf.CfgKey("LOG_DIR")
		assert.Equal(t, "tmp/log", val)
		assert.NoError(t, err)

		val, err = cpf.CfgKey("ABS_DIR")
		assert.Equal(t, "/some/absolute/path/log", val)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		cpf := FakeLoad(tspy, "not_existing_file")

		// --- Then ---
		assert.Nil(t, cpf)
	})

	t.Run("invalid config file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("dotenv: invalid config line: 1")
		tspy.Close()

		// --- When ---
		cpf := FakeLoad(tspy, "testdata/invalid.conf")

		// --- Then ---
		assert.Nil(t, cpf)
	})
}

func Test_Fake_Reader(t *testing.T) {
	// --- When ---
	cpf := FakeLoad(t, "../testdata/example0.conf")

	// --- When ---
	rdr := cpf.Reader()

	// --- Then ---
	want := `ABS_DIR="/some/absolute/path/log"
LOG_DIR="tmp/log"
VALUE=0`
	assert.Equal(t, want, iokit.ReadAllStr(t, rdr))
}

func Test_Fake_Set(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		tt := []struct {
			testN string

			val  any
			want string
		}{
			{"1", "abc", "abc"},
			{"2", 1, "1"},
			{"3", float32(1.23), "1.23"},
			{"4", 1.23, "1.23"},
			{"5", time.Second, "1s"},
			{"6", time.UTC, "UTC"},
			{
				"7",
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
				val, err := cpf.CfgKey("KEY")
				assert.NoError(t, err)
				assert.Equal(t, tc.want, val)
			})
		}
	})

	t.Run("fatal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "unsupported config type %#v"
		tspy.ExpectLogEqual(wMsg, uint(10))
		tspy.Close()

		// --- When ---
		_, _ = NewFake(tspy).Set("KEY", uint(10)).CfgKey("KEY")
	})
}

func Test_Fake_SetFrom(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	cpf := NewFake(tspy)
	cpf.Set("KEY0", "VAL0")

	// --- When ---
	m := map[string]string{
		"KEY0": "VAL99",
		"KEY1": "VAL1",
	}
	cpf.SetFrom(m)

	// --- Then ---
	val, err := cpf.CfgKey("KEY0")
	assert.NoError(t, err)
	assert.Equal(t, "VAL99", val)

	val, err = cpf.CfgKey("KEY1")
	assert.NoError(t, err)
	assert.Equal(t, "VAL1", val)
}

func Test_Fake_SetFromWithPrefix(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("KEY0", "VAL0")

		// --- When ---
		m := map[string]string{
			"KEY0": "VAL99",
			"KEY1": "VAL1",
		}
		cpf.SetFromWithPrefix("", m)

		// --- Then ---
		val, err := cpf.CfgKey("KEY0")
		assert.NoError(t, err)
		assert.Equal(t, "VAL99", val)

		val, err = cpf.CfgKey("KEY1")
		assert.NoError(t, err)
		assert.Equal(t, "VAL1", val)
	})

	t.Run("with prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("KEY0", "VAL0")

		// --- When ---
		m := map[string]string{
			"KEY0": "VAL99",
			"KEY1": "VAL1",
		}
		cpf.SetFromWithPrefix("PREF", m)

		// --- Then ---
		val, err := cpf.CfgKey("PREF_KEY0")
		assert.NoError(t, err)
		assert.Equal(t, "VAL99", val)

		val, err = cpf.CfgKey("PREF_KEY1")
		assert.NoError(t, err)
		assert.Equal(t, "VAL1", val)
	})
}

func Test_Fake_DeleteKey(t *testing.T) {
	// --- Given ---
	cpf := FakeLoad(t, "../testdata/example0.conf")

	// --- When ---
	cpf.DeleteKey("LOG_DIR")

	// --- Then ---
	val, err := cpf.CfgKey("LOG_DIR")
	xrrtest.AssertFieldIs(t, "LOG_DIR", xcfg.ErrKeyMissing, err)
	assert.Equal(t, "", val)
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
	cpf := FakeLoad(t, "../testdata/example0.conf")

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

func Test_Fake_MustKey(t *testing.T) {
	// --- When ---
	cpf := FakeLoad(t, "../testdata/example0.conf").MustKey("VALUE")

	// --- Then ---
	assert.Equal(t, "0", cpf)
}

func Test_Fake_MustKey_fatal(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectError()
	tspy.ExpectLogEqual("NOT_EXISTING: config key missing")
	tspy.Close()

	// --- When ---
	cpf := FakeLoad(tspy, "../testdata/example0.conf").MustKey("NOT_EXISTING")

	// --- Then ---
	assert.Equal(t, "", cpf)
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

		want := `LOG_DIR="/var/log"
VALUE="value"`
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))
	})

	t.Run("path is a directory error", func(t *testing.T) {
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

	t.Run("invalid argument", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("setenv: invalid argument")
		tspy.Close()

		cpf := NewFake(tspy)
		cpf.Set("", "value")

		// --- When ---
		cpf.SetEnv()
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
			"time.Time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"2000-01-02T03:04:05Z",
		},
		{"time.Location", time.UTC, "UTC"},
		{"fmt.Stringer", TStringer(123), "123"},
		{"bool", true, "true"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			// --- When ---
			have := anyToStr(tspy, tc.in)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_anyToStr(t *testing.T) {
	t.Run("not supported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("unsupported config type xcfgtest.Abc{}")
		tspy.Close()

		type Abc struct{}

		// --- When ---
		have := anyToStr(tspy, Abc{})

		// --- Then ---
		assert.Empty(t, have)
	})
}
