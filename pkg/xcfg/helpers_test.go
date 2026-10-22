// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testkit/pkg/iokit"
	"github.com/ctx42/testkit/pkg/oskit"
)

func Test_Open(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		have, err := Open("testdata/example0.conf")

		// --- Then ---
		assert.NoError(t, err)
		assert.NotNil(t, have)
	})

	t.Run("error - not existing", func(t *testing.T) {
		// --- When ---
		have, err := Open("/some/not/existing/path/app.conf")

		// --- Then ---
		assert.Error(t, err)
		assert.ErrorContain(t, "no such file or directory", err)
		assert.Nil(t, have)
	})

	t.Run("error - directory", func(t *testing.T) {
		// --- When ---
		have, err := Open("testdata")

		// --- Then ---
		assert.Error(t, err)
		assert.ErrorContain(t, "read testdata: is a directory", err)
		assert.Nil(t, have)
	})
}

func Test_Merge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dst := make(map[string]string, 5)
		dst["KEY3"] = "VAL3"

		src0 := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
		}
		src1 := map[string]string{
			"KEY4": "VAL4",
			"KEY5": "VAL5",
		}

		// --- When ---
		have := Merge(dst, src0, src1)

		// --- Then ---
		want := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
			"KEY3": "VAL3",
			"KEY4": "VAL4",
			"KEY5": "VAL5",
		}
		assert.Equal(t, want, have)
		assert.Same(t, dst, have)
	})

	t.Run("dst argument is nil", func(t *testing.T) {
		// --- Given ---
		src0 := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
		}
		src1 := map[string]string{
			"KEY4": "VAL4",
			"KEY5": "VAL5",
		}

		// --- When ---
		have := Merge(nil, src0, src1)

		// --- Then ---
		want := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
			"KEY4": "VAL4",
			"KEY5": "VAL5",
		}
		assert.Equal(t, want, have)
	})

	t.Run("src is nil", func(t *testing.T) {
		// --- Given ---
		dst := make(map[string]string, 5)
		dst["KEY3"] = "VAL3"

		// --- When ---
		have := Merge(dst, nil)

		// --- Then ---
		assert.Equal(t, map[string]string{"KEY3": "VAL3"}, have)
		assert.Same(t, dst, have)
	})

	t.Run("one of the sources is nil", func(t *testing.T) {
		// --- Given ---
		dst := make(map[string]string, 5)
		dst["KEY3"] = "VAL3"

		src0 := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
		}

		// --- When ---
		have := Merge(dst, src0, nil)

		// --- Then ---
		want := map[string]string{
			"KEY1": "VAL1",
			"KEY2": "VAL2",
			"KEY3": "VAL3",
		}
		assert.Equal(t, want, have)
		assert.Same(t, dst, have)
	})

	t.Run("later source wins", func(t *testing.T) {
		// --- Given ---
		dst := map[string]string{"KEY": "DST"}
		src0 := map[string]string{"KEY": "SRC0"}
		src1 := map[string]string{"KEY": "SRC1"}

		// --- When ---
		have := Merge(dst, src0, src1)

		// --- Then ---
		assert.Equal(t, map[string]string{"KEY": "SRC1"}, have)
	})
}

func Test_DirExists_tabular(t *testing.T) {
	tt := []struct {
		testN string

		path string
		want bool
	}{
		{"file", "testdata/equals.conf", false},
		{"not existing", "testdata/not_existing.conf", false},
		{"directory", "testdata", true},
		{"link to file", "testdata/link.conf", false},
		{"under a file", "testdata/equals.conf/sub", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := DirExists(tc.path)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_doubleQuoteEscape_tabular(t *testing.T) {
	tt := []struct {
		testN string

		in   string
		want string
	}{
		{"empty", "", ""},
		{"backslash", "\\", "\\\\"},
		{"new line", "\n", "\\n"},
		{"carriage return", "\r", "\\r"},
		{"exclamation", "!", "\\!"},
		{"dollar", "$", "\\$"},
		{"double quote", `"`, "\\\""},
		{"backtick", "`", "\\`"},
		{"text with new line", "abc\ndef", "abc\\ndef"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := doubleQuoteEscape(tc.in)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_isDotenvKey_tabular(t *testing.T) {
	tt := []struct {
		testN string

		key  string
		want bool
	}{
		{"upper case", "KEY_NAME", true},
		{"lower case with dot", "key.name", true},
		{"dash", "KEY-NAME", true},
		{"non-ASCII", "KLUCZ_Ż", true},
		{"empty", "", false},
		{"equals", "A=B", false},
		{"colon", "A:B", false},
		{"hash", "#A", false},
		{"space", "A B", false},
		{"tab", "A\tB", false},
		{"new line", "A\nB", false},
		{"carriage return", "A\rB", false},
		{"no-break space", "\u00a0A", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := isDotenvKey(tc.key)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_maybeDecode(t *testing.T) {
	t.Run("not encoded config", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/quoted.conf"
		r := oskit.Open(t, pth)

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("input shorter than prefix", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("A=B")

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "A=B", iokit.ReadAllStr(t, have))
	})

	t.Run("empty input", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("")

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "", iokit.ReadAllStr(t, have))
	})

	t.Run("simple encoded config", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("base64:QT1CCkM9RAo")

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "A=B\nC=D\n", iokit.ReadAllStr(t, have))
	})

	t.Run("encoded config", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/quoted.conf"
		enc := base64.RawURLEncoding.EncodeToString(oskit.ReadFile(t, pth))
		r := strings.NewReader("base64:" + enc)

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("error - invalid base64", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("base64:%%%")

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		wMsg := "decode base64: illegal base64 data at input byte 0"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})
}

func Test_maybeDecode_tabular(t *testing.T) {
	tt := []struct {
		testN string

		in   string
		want string
	}{
		{"padded", "base64:QT1CCkM9RA==", "A=B\nC=D"},
		{"trailing new line", "base64:QT1CCkM9RAo\n", "A=B\nC=D\n"},
		{"trailing CRLF", "base64:QT1CCkM9RAo\r\n", "A=B\nC=D\n"},
		{"trailing space", "base64:QT1CCkM9RAo ", "A=B\nC=D\n"},
		{"leading space", "base64: QT1CCkM9RAo", "A=B\nC=D\n"},
		{"padded with new line", "base64:QT1CCkM9RA==\n", "A=B\nC=D"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := maybeDecode(strings.NewReader(tc.in))

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, iokit.ReadAllStr(t, have))
		})
	}
}

func Test_EnvSplit_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		want map[string]string
	}{
		{"empty", []string{}, map[string]string{}},
		{"empty entry", []string{""}, map[string]string{}},
		{"key value", []string{"A=B"}, map[string]string{"A": "B"}},
		{"equals in value", []string{"A=B=C"}, map[string]string{"A": "B=C"}},
		{"empty value", []string{"A="}, map[string]string{"A": ""}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := EnvSplit(tc.env)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}
