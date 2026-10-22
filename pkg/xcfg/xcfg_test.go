// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg_test

import (
	"bytes"
	"encoding/base64"
	"syscall"
	"testing"

	"github.com/ctx42/dotenv/pkg/dotenv"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xcfg/pkg/xcfg/xcfgtest"

	. "github.com/ctx42/xcfg/pkg/xcfg"
)

func Test_New(t *testing.T) {
	t.Run("error - read failure", func(t *testing.T) {
		// --- Given ---
		fil := oskit.Open(t, "testdata/example0.conf")
		er := iokit.ErrReader(fil, 10)

		// --- When ---
		have, err := New(er)

		// --- Then ---
		assert.ErrorEqual(t, "config: read error", err)
		assert.ErrorIs(t, iokit.ErrRead, err)
		assert.Nil(t, have)
	})

	t.Run("parse simple", func(t *testing.T) {
		// --- Given ---
		rdr := bytes.NewReader([]byte("ONE=1\nTWO='2'\nTHREE = \"3\""))

		// --- When ---
		have, err := New(rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"ONE":   "1",
			"TWO":   "2",
			"THREE": "3",
		}
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse plain", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/plain.conf"))

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
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse exported", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/exported.conf"))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "2",
			"OPTION_B": "\\n",
		}
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse equals", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/equals.conf"))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "postgres://localhost:5432/database?sslmode=disable",
		}
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse quoted", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/quoted.conf"))

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
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse substitutions", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/substitutions.conf"))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "1",
			"OPTION_C": "1",
			"OPTION_D": "11",
			"OPTION_E": "",
		}
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("parse base64 encoded string", func(t *testing.T) {
		// --- Given ---
		rdr := bytes.NewReader([]byte("base64:QT1CCkM9RAo"))

		// --- When ---
		have, err := New(rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"A": "B",
			"C": "D",
		}
		assert.Equal(t, Config(want), have)
		assert.Len(t, len(want), have)
	})

	t.Run("error - invalid line", func(t *testing.T) {
		// --- When ---
		have, err := New(xcfgtest.Open(t, "testdata/invalid1.conf"))

		// --- Then ---
		assert.ErrorIs(t, dotenv.ErrInvLine, err)
		assert.Nil(t, have)
	})
}

func Test_NewFromPath(t *testing.T) {
	t.Run("error - empty path", func(t *testing.T) {
		// --- When ---
		have, err := NewFromPath("")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, have)
	})

	t.Run("error - empty path with env", func(t *testing.T) {
		// --- Given ---
		env := []string{"A=B"}

		// --- When ---
		have, err := NewFromPath("", WithEnv(env))

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, have)
	})

	t.Run("error - not existing path", func(t *testing.T) {
		// --- When ---
		have, err := NewFromPath("testdata/not_existing.conf")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, have)
	})

	t.Run("error - not existing path with env", func(t *testing.T) {
		// --- Given ---
		env := []string{"A=B"}

		// --- When ---
		have, err := NewFromPath("testdata/not_existing.conf", WithEnv(env))

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.Nil(t, have)
	})

	t.Run("error - path is a directory", func(t *testing.T) {
		// --- When ---
		have, err := NewFromPath("testdata")

		// --- Then ---
		assert.ErrorIs(t, ErrCfgFileNotFound, err)
		assert.ErrorIs(t, syscall.EISDIR, err)
		xrrtest.AssertCode(t, ECCfgFileNotFound, err)
		wMsg := "config file not found; read testdata: is a directory"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("existing path", func(t *testing.T) {
		// --- When ---
		have, err := NewFromPath("testdata/example1.conf")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Config{"VALUE": "1"}, have)
	})

	t.Run("env overrides matching keys", func(t *testing.T) {
		// --- Given ---
		env := []string{"VALUE=42", "OTHER=1"}

		// --- When ---
		have, err := NewFromPath("testdata/example1.conf", WithEnv(env))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Config{"VALUE": "42"}, have)
	})

	t.Run("error - invalid config", func(t *testing.T) {
		// --- When ---
		have, err := NewFromPath("testdata/invalid1.conf")

		// --- Then ---
		assert.ErrorEqual(t, "config: dotenv: invalid config line: 1", err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid config with env", func(t *testing.T) {
		// --- Given ---
		env := []string{"A=B"}

		// --- When ---
		have, err := NewFromPath("testdata/invalid1.conf", WithEnv(env))

		// --- Then ---
		assert.ErrorEqual(t, "config: dotenv: invalid config line: 1", err)
		assert.Nil(t, have)
	})

	t.Run("options built as a list", func(t *testing.T) {
		// --- Given ---
		var opts []Option
		opts = append(opts, WithEnv([]string{"VALUE=42"}))

		// --- When ---
		have, err := NewFromPath("testdata/example1.conf", opts...)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Config{"VALUE": "42"}, have)
	})

	t.Run("encoded path updated by env", func(t *testing.T) {
		// --- Given ---
		enc := base64.RawURLEncoding.EncodeToString([]byte("A=B\nC=D\nE=F"))
		env := []string{"A=G"}

		// --- When ---
		have, err := NewFromPath("base64:"+enc, WithEnv(env))

		// --- Then ---
		assert.NoError(t, err)
		want := Config{
			"A": "G",
			"C": "D",
			"E": "F",
		}
		assert.Equal(t, want, have)
	})
}

func Test_NewFromEnv(t *testing.T) {
	// --- Given ---
	env := []string{"A=B", "C=D"}

	// --- When ---
	have := NewFromEnv(env)

	// --- Then ---
	assert.Equal(t, Config{"A": "B", "C": "D"}, have)
}

func Test_Config_CfgKey(t *testing.T) {
	t.Run("get existing", func(t *testing.T) {
		// --- Given ---
		cfg := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cfg.CfgKey("LOG_DIR")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "tmp/log", have)
	})

	t.Run("get existing zero value", func(t *testing.T) {
		// --- Given ---
		cfg := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cfg.CfgKey("VALUE")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "0", have)
	})

	t.Run("error - missing key", func(t *testing.T) {
		// --- Given ---
		cfg := xcfgtest.NewFromPath(t, "testdata/example0.conf")

		// --- When ---
		have, err := cfg.CfgKey("NOT_EXISTING")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "NOT_EXISTING", ErrKeyMissing, err)
		assert.ErrorEqual(t, "NOT_EXISTING: config key missing", err)
		assert.Equal(t, "", have)
	})

	t.Run("env does not override file", func(t *testing.T) {
		// --- Given ---
		cfg := xcfgtest.NewFromPath(t, "testdata/example0.conf")
		t.Setenv("VALUE", "2")

		// --- When ---
		have, err := cfg.CfgKey("VALUE")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "0", have)
	})

	t.Run("error - key only in env", func(t *testing.T) {
		// --- Given ---
		cfg := xcfgtest.NewFromPath(t, "testdata/example0.conf")
		t.Setenv("VALUE_ENV", "123")

		// --- When ---
		have, err := cfg.CfgKey("VALUE_ENV")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "VALUE_ENV", ErrKeyMissing, err)
		assert.ErrorEqual(t, "VALUE_ENV: config key missing", err)
		assert.Equal(t, "", have)
	})
}

func Test_Config_Marshal(t *testing.T) {
	t.Run("error - invalid keys", func(t *testing.T) {
		// --- Given ---
		cfg := Config{"A B": "1", "A=B": "2", "OK": "3"}

		// --- When ---
		have, err := cfg.Marshal()

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertFieldIs(t, "A B", ErrKeyInvalid, err)
		xrrtest.AssertFieldIs(t, "A=B", ErrKeyInvalid, err)
		xrrtest.AssertFieldCode(t, "A=B", ECCfgKeyInvValue, err)
		assert.Empty(t, have)
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
			have, err := cfg.Marshal()

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Config_Encode(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{
			"LOG_DIR": "/var/log",
			"VALUE":   "value",
		}
		cfg := Config(m)

		// --- When ---
		have, err := cfg.Encode()

		// --- Then ---
		assert.NoError(t, err)
		want := "base64:TE9HX0RJUj0iL3Zhci9sb2ciClZBTFVFPSJ2YWx1ZSI"
		assert.Equal(t, want, have)

		check := must.Value(base64.RawURLEncoding.DecodeString(want[7:]))
		want = "" +
			"LOG_DIR=\"/var/log\"\n" +
			"VALUE=\"value\""
		assert.Equal(t, want, string(check))
	})

	t.Run("error - invalid key", func(t *testing.T) {
		// --- When ---
		have, err := Config{"A B": "1"}.Encode()

		// --- Then ---
		xrrtest.AssertFieldIs(t, "A B", ErrKeyInvalid, err)
		assert.Empty(t, have)
	})
}

func Test_Config_SetFrom(t *testing.T) {
	t.Run("set", func(t *testing.T) {
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

		assert.Equal(t, "VAL99", must.Value(cfg.CfgKey("KEY0")))

		assert.Equal(t, "VAL1", must.Value(cfg.CfgKey("KEY1")))
	})

	t.Run("nil config", func(t *testing.T) {
		// --- Given ---
		var cfg Config

		// --- When ---
		have := cfg.SetFrom(map[string]string{"KEY": "VAL"})

		// --- Then ---
		assert.Equal(t, Config{"KEY": "VAL"}, have)
	})
}

func Test_Unmarshal(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		// --- When ---
		have, err := Unmarshal("A=B\nC='D'")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Config{"A": "B", "C": "D"}, have)
	})

	t.Run("marshaled escapes", func(t *testing.T) {
		// --- Given ---
		cfg := Config{"K": "a\\b\r\n!$`\"' #c"}

		// --- When ---
		have, err := Unmarshal(must.Value(cfg.Marshal()))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, cfg, have)
	})

	t.Run("error - invalid line", func(t *testing.T) {
		// --- When ---
		have, err := Unmarshal("INVALID LINE")

		// --- Then ---
		assert.ErrorIs(t, dotenv.ErrInvLine, err)
		assert.Nil(t, have)
	})
}

func Test_Unmarshal_tabular(t *testing.T) {
	tt := []struct {
		testN string

		fixture string
	}{
		{"equals", "testdata/equals.conf"},
		{"exported", "testdata/exported.conf"},
		{"plain", "testdata/plain.conf"},
		{"quoted", "testdata/quoted.conf"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cfg := must.Value(New(xcfgtest.Open(t, tc.fixture)))

			// --- When ---
			have, err := Unmarshal(must.Value(cfg.Marshal()))

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, cfg, have)
		})
	}
}
