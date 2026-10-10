// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"syscall"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xcfg/pkg/xcfg/xcfgtest"

	. "github.com/ctx42/xcfg/pkg/xcfg"
)

func Test_New(t *testing.T) {
	t.Run("readr error", func(t *testing.T) {
		// --- Given ---
		fil := oskit.Open(t, "testdata/example0.conf")
		er := iokit.ErrReader(fil, 10)

		// --- When ---
		cfg, err := New(er)

		// --- Then ---
		assert.ErrorEqual(t, "config: read error", err)
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Nil(t, cfg)
	})

	t.Run("cast map", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"KEY_1": "VAL_1",
			"KEY_2": "VAL_2",
			"PTH":   "testdata",
		}

		// --- When ---
		cfg := Config(data)

		// --- Then ---
		value, err := cfg.CfgKey("KEY_1")
		assert.Equal(t, "VAL_1", value)
		assert.NoError(t, err)

		value, err = cfg.CfgKey("KEY_2")
		assert.Equal(t, "VAL_2", value)
		assert.NoError(t, err)

		value, err = cfg.CfgKey("PTH")
		assert.Equal(t, "testdata", value)
		assert.NoError(t, err)

		value, err = cfg.CfgKey("NOT_EXISTING")
		assert.Equal(t, "", value)
		xrrtest.AssertFieldIs(t, "NOT_EXISTING", ErrKeyMissing, err)
		assert.ErrorEqual(t, "NOT_EXISTING: config key missing", err)

		want := "testdata"
		value, err = cfg.CfgKey("PTH")
		assert.Equal(t, want, value)
		assert.NoError(t, err)
	})

	t.Run("parse simple", func(t *testing.T) {
		// --- Given ---
		rdr := bytes.NewReader([]byte("ONE=1\nTWO='2'\nTHREE = \"3\""))

		// --- When ---
		cfg, err := New(rdr)

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"ONE":   "1",
			"TWO":   "2",
			"THREE": "3",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse plain", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/plain.conf"))

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "3",
			"OPTION_D": "4",
			"OPTION_E": "5",
			"OPTION_F": "",
			"OPTION_G": "",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse exported", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/exported.conf"))

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"OPTION_A": "2",
			"OPTION_B": "\\n",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse equals", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/equals.conf"))

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"OPTION_A": "postgres://localhost:5432/database?sslmode=disable",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse quoted", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/quoted.conf"))

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "",
			"OPTION_D": "\\n",
			"OPTION_E": "1",
			"OPTION_F": "2",
			"OPTION_G": "",
			"OPTION_H": "\n",
			"OPTION_I": "echo 'asd'",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse substitutions", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/substitutions.conf"))

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "1",
			"OPTION_C": "1",
			"OPTION_D": "11",
			"OPTION_E": "",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse base64 encoded string", func(t *testing.T) {
		// --- Given ---
		rdr := bytes.NewReader([]byte("base64:QT1CCkM9RAo"))

		// --- When ---
		cfg, err := New(rdr)

		// --- Then ---
		assert.NoError(t, err)

		want := map[string]string{
			"A": "B",
			"C": "D",
		}
		assert.Equal(t, Config(want), cfg)
		assert.Len(t, len(want), cfg)
	})

	t.Run("parse invalid file", func(t *testing.T) {
		// --- When ---
		cfg, err := New(xcfgtest.Open(t, "testdata/invalid1.conf"))

		// --- Then ---
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func Test_NewFromPath(t *testing.T) {
	t.Run("empty path nil env", func(t *testing.T) {
		// --- When ---
		cfg, err := NewFromPath("")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, cfg)
	})

	t.Run("empty path env set", func(t *testing.T) {
		// --- When ---
		env := []string{"A=B"}
		cfg, err := NewFromPath("", WithEnv(env))

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, cfg)
	})

	t.Run("not existing path nil env", func(t *testing.T) {
		// --- When ---
		cfg, err := NewFromPath("testdata/not_existing.conf")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, cfg)
	})

	t.Run("not existing path env set", func(t *testing.T) {
		// --- When ---
		env := []string{"A=B"}
		cfg, err := NewFromPath("testdata/not_existing.conf", WithEnv(env))

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, cfg)
	})

	t.Run("error - path is a directory", func(t *testing.T) {
		// --- When ---
		cfg, err := NewFromPath("testdata")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.ErrorIs(t, syscall.EISDIR, err)
		xrrtest.AssertCode(t, ECCfgFileNotFound, err)
		wMsg := "config file not found; read testdata: is a directory"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, cfg)
	})

	t.Run("existing path nil env", func(t *testing.T) {
		// --- When ---
		cfg, err := NewFromPath("testdata/example1.conf")

		// --- Then ---
		assert.NoError(t, err)
		value, err := cfg.CfgKey("VALUE")
		assert.NoError(t, err)
		assert.Equal(t, "1", value)
	})

	t.Run("matching keys overwritten from environment", func(t *testing.T) {
		// --- When ---
		env := []string{"VALUE=42"}
		cfg, err := NewFromPath("testdata/example1.conf", WithEnv(env))

		// --- Then ---
		assert.NoError(t, err)
		value, err := cfg.CfgKey("VALUE")
		assert.NoError(t, err)
		assert.Equal(t, "42", value)
	})

	t.Run("invalid config nil env", func(t *testing.T) {
		// --- When ---
		cfg, err := NewFromPath("testdata/invalid1.conf")

		// --- Then ---
		assert.ErrorEqual(t, "config: dotenv: invalid config line: 1", err)
		assert.Nil(t, cfg)
	})

	t.Run("invalid config env set", func(t *testing.T) {
		// --- When ---
		env := []string{"A=B"}
		cfg, err := NewFromPath("testdata/invalid1.conf", WithEnv(env))

		// --- Then ---
		assert.ErrorEqual(t, "config: dotenv: invalid config line: 1", err)
		assert.Nil(t, cfg)
	})

	t.Run("encoded path updated by env", func(t *testing.T) {
		// --- When ---
		enc := base64.RawURLEncoding.EncodeToString([]byte("A=B\nC=D\nE=F"))
		pth := "base64:" + enc
		env := []string{"A=G"}
		cfg, err := NewFromPath(pth, WithEnv(env))

		// --- Then ---
		assert.NoError(t, err)
		want := Config{
			"A": "G",
			"C": "D",
			"E": "F",
		}
		assert.Equal(t, want, cfg)
	})
}

func Test_NewFromEnv(t *testing.T) {
	// --- Given ---
	env := []string{"A=B", "C=D"}

	// --- When ---
	cfg := NewFromEnv(env)

	// --- Then ---
	val, err := cfg.CfgKey("A")
	assert.NoError(t, err)
	assert.Equal(t, "B", val)

	val, err = cfg.CfgKey("E")
	xrrtest.AssertFieldIs(t, "E", ErrKeyMissing, err)
	assert.ErrorEqual(t, "E: config key missing", err)
	assert.Equal(t, "", val)
}

func Test_Config_CfgKey(t *testing.T) {
	t.Run("get existing", func(t *testing.T) {
		// --- Given ---
		cpv := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cpv.CfgKey("LOG_DIR")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "tmp/log", have)
	})

	t.Run("get existing zero value", func(t *testing.T) {
		// --- Given ---
		cpv := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cpv.CfgKey("VALUE")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "0", have)
	})

	t.Run("get not existing", func(t *testing.T) {
		// --- Given ---
		cpv := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cpv.CfgKey("NOT_EXISTING")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "NOT_EXISTING", ErrKeyMissing, err)
		assert.ErrorEqual(t, "NOT_EXISTING: config key missing", err)
		assert.Equal(t, "", have)
	})

	t.Run("env does not override file", func(t *testing.T) {
		// --- Given ---
		cpv := xcfgtest.NewFromPath(t, "testdata/example0.conf")
		t.Setenv("VALUE", "2")

		// --- When ---
		have, err := cpv.CfgKey("VALUE")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "0", have)
	})

	t.Run("key not in file but in env", func(t *testing.T) {
		// --- Given ---
		cpv := xcfgtest.NewFromPath(t, "testdata/example0.conf")
		t.Setenv("VALUE_ENV", "123")

		// --- When ---
		have, err := cpv.CfgKey("VALUE_ENV")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "VALUE_ENV", ErrKeyMissing, err)
		assert.ErrorEqual(t, "VALUE_ENV: config key missing", err)
		assert.Equal(t, "", have)
	})
}

func Test_Config_Marshal_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  string
		want string
	}{
		{"integer", "42", "K=42"},
		{"negative integer", "-42", "K=-42"},
		{"leading zero", "0755", `K="0755"`},
		{"plus sign", "+1", `K="+1"`},
		{"string", "abc", `K="abc"`},
		{"escaped", "a\"$\n", `K="a\"\$\n"`},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cfg := Config{"K": tc.val}

			// --- When ---
			have := cfg.Marshal()

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Config_Unmarshal_Marshal(t *testing.T) {
	tt := []struct {
		testN string

		fixture string
	}{
		{"1", "testdata/equals.conf"},
		{"2", "testdata/exported.conf"},
		{"3", "testdata/plain.conf"},
		{"4", "testdata/quoted.conf"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cfg, err := New(xcfgtest.Open(t, tc.fixture))
			assert.NoError(t, err)

			// --- When ---
			have, err := Unmarshal(cfg.Marshal())

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, cfg, have)
		})
	}
}

func Test_Config_Encode(t *testing.T) {
	// --- Given ---
	m := map[string]string{
		"LOG_DIR": "/var/log",
		"VALUE":   "value",
	}
	cfg := Config(m)

	// --- When ---
	have := cfg.Encode()

	// --- Then ---
	want := "base64:TE9HX0RJUj0iL3Zhci9sb2ciClZBTFVFPSJ2YWx1ZSI"
	assert.Equal(t, want, have)

	check := must.Value(base64.RawURLEncoding.DecodeString(want[7:]))
	want = "" +
		"LOG_DIR=\"/var/log\"\n" +
		"VALUE=\"value\""
	assert.Equal(t, want, string(check))
}

func Test_Config_SetFrom(t *testing.T) {
	// --- Given ---
	cfg := Config(map[string]string{"KEY0": "VAL0"})
	m := map[string]string{
		"KEY0": "VAL99",
		"KEY1": "VAL1",
	}

	// --- When ---
	have := cfg.SetFrom(m)

	// --- Then ---
	assert.Same(t, cfg, have)

	val, err := cfg.CfgKey("KEY0")
	assert.NoError(t, err)
	assert.Equal(t, "VAL99", val)

	val, err = cfg.CfgKey("KEY1")
	assert.NoError(t, err)
	assert.Equal(t, "VAL1", val)
}

func Test_maybeDecode(t *testing.T) {
	t.Run("not encoded config", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/quoted.conf"
		r := xcfgtest.Open(t, pth)

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("input shorter than prefix", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("A=B")

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "A=B", iokit.ReadAllStr(t, have))
	})

	t.Run("empty input", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("")

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "", iokit.ReadAllStr(t, have))
	})

	t.Run("simple encoded config", func(t *testing.T) {
		// --- Given ---
		src := "A=B\nC=D\n"
		enc := "base64:QT1CCkM9RAo"
		r := strings.NewReader(enc)

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, src, iokit.ReadAllStr(t, have))
	})

	t.Run("encoded config", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/quoted.conf"
		enc := base64.RawURLEncoding.EncodeToString(oskit.ReadFile(t, pth))
		r := strings.NewReader("base64:" + enc)

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("decode error", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("base64:%%%")

		// --- When ---
		have, err := MaybeDecode(r)

		// --- Then ---
		assert.ErrorEqual(t, "illegal base64 data at input byte 0", err)
		assert.Nil(t, have)
	})
}
