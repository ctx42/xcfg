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

	t.Run("not existing", func(t *testing.T) {
		// --- When ---
		have, err := Open("/some/not/existing/path/app.conf")

		// --- Then ---
		assert.Error(t, err)
		assert.ErrorContain(t, "no such file or directory", err)
		assert.Nil(t, have)
	})

	t.Run("directory", func(t *testing.T) {
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
		assert.Len(t, 5, dst)
		assert.HasKey(t, "KEY1", dst)
		assert.HasKey(t, "KEY2", dst)
		assert.HasKey(t, "KEY3", dst)
		assert.HasKey(t, "KEY4", dst)
		assert.HasKey(t, "KEY5", dst)

		assert.Len(t, 5, have)
		assert.HasKey(t, "KEY1", have)
		assert.HasKey(t, "KEY2", have)
		assert.HasKey(t, "KEY3", have)
		assert.HasKey(t, "KEY4", have)
		assert.HasKey(t, "KEY5", have)
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
		assert.Len(t, 4, have)
		assert.HasKey(t, "KEY1", have)
		assert.HasKey(t, "KEY2", have)
		assert.HasKey(t, "KEY4", have)
		assert.HasKey(t, "KEY5", have)
	})

	t.Run("src is nil", func(t *testing.T) {
		// --- Given ---
		dst := make(map[string]string, 5)
		dst["KEY3"] = "VAL3"

		// --- When ---
		have := Merge(dst, nil)

		// --- Then ---
		assert.Len(t, 1, dst)
		assert.HasKey(t, "KEY3", dst)

		assert.Len(t, 1, have)
		assert.HasKey(t, "KEY3", have)
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
		assert.Len(t, 3, dst)
		assert.HasKey(t, "KEY1", dst)
		assert.HasKey(t, "KEY2", dst)
		assert.HasKey(t, "KEY3", dst)

		assert.Len(t, 3, have)
		assert.HasKey(t, "KEY1", have)
		assert.HasKey(t, "KEY2", have)
		assert.HasKey(t, "KEY3", have)
	})
}

func Test_DirExists_tabular(t *testing.T) {
	tt := []struct {
		testN string

		path string
		want bool
	}{
		{"1", "testdata/equals.conf", false}, // A file is not a directory.
		{"2", "testdata/not_existing.conf", false},
		{"3", "testdata", true},
		{"4", "testdata/link.conf", false},
		{"5", "testdata/equals.conf/sub", false}, // Under a file.
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
		{"1", "\\", "\\\\"},
		{"2", "\n", "\\n"},
		{"3", "\r", "\\r"},
		{"4", "!", "\\!"},
		{"5", "$", "\\$"},
		{"6", `"`, "\\\""},
		{"7", "abc\ndef", "abc\\ndef"},
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
		src := "A=B\nC=D\n"
		enc := "base64:QT1CCkM9RAo"
		r := strings.NewReader(enc)

		// --- When ---
		have, err := maybeDecode(r)

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
		have, err := maybeDecode(r)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, oskit.ReadFileStr(t, pth), iokit.ReadAllStr(t, have))
	})

	t.Run("decode error", func(t *testing.T) {
		// --- Given ---
		r := strings.NewReader("base64:%%%")

		// --- When ---
		have, err := maybeDecode(r)

		// --- Then ---
		assert.ErrorEqual(t, "illegal base64 data at input byte 0", err)
		assert.Nil(t, have)
	})
}

func Test_EnvSplit_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		want map[string]string
	}{
		{"1", []string{}, map[string]string{}},
		{"1a", []string{""}, map[string]string{}},
		{"2", []string{"A=B"}, map[string]string{"A": "B"}},
		{"3", []string{"A=B=C"}, map[string]string{"A": "B=C"}},
		{"4", []string{"A="}, map[string]string{"A": ""}},
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
