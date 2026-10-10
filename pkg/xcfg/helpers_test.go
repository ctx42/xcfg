// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
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

func Test_GetOr_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		key  string
		def  string
		want string
	}{
		{"key exists and has value", []string{"KEY=VAL"}, "KEY", "DEF", "VAL"},
		{"key exists and is empty", []string{"KEY="}, "KEY", "DEF", ""},
		{"key does not exist", []string{"KEY=VAL"}, "OTHER", "DEF", "DEF"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cpv := NewFromEnv(tc.env)

			// --- When ---
			have := GetOr(tc.key, cpv, tc.def)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Prefix_tabular(t *testing.T) {
	tt := []struct {
		testN string

		prefix string
		key    string
		want   string
	}{
		{"empty prefix", "", "KEY_NAME", "KEY_NAME"},
		{"with prefix", "PREFIX", "KEY_NAME", "PREFIX_KEY_NAME"},
		{"empty key", "PREFIX", "", "PREFIX"},
		{"prefix and key empty", "", "", ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Equal(t, tc.want, Prefix(tc.prefix, tc.key))
		})
	}
}

func Test_StripPrefix_tabular(t *testing.T) {
	tt := []struct {
		testN string

		prefix string
		key    string
		want   string
	}{
		{"empty prefix", "", "KEY_NAME", "KEY_NAME"},
		{"with prefix", "PREFIX", "PREFIX_KEY_NAME", "KEY_NAME"},
		{"empty key", "PREFIX", "", ""},
		{"prefix and key empty", "", "", ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Equal(t, tc.want, StripPrefix(tc.prefix, tc.key))
		})
	}
}

func Test_StripPrefixes(t *testing.T) {
	t.Run("strip", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"MY_KEY_3": "VAL_3",
		}

		// --- When ---
		have := StripPrefixes("MY", data)

		// --- Then ---
		want := map[string]string{
			"KEY_1": "VAL_1",
			"KEY_2": "VAL_2",
			"KEY_3": "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("empty prefix", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"MY_KEY_3": "VAL_3",
		}

		// --- When ---
		have := StripPrefixes("", data)

		// --- Then ---
		want := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"MY_KEY_3": "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		have := StripPrefixes("", nil)

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_AddPrefixes(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}

		// --- When ---
		have := AddPrefixes("MY", data)

		// --- Then ---
		want := map[string]string{
			"MY_MY_KEY_1": "VAL_1",
			"MY_KEY_2":    "VAL_2",
			"MY_KEY_3":    "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("empty prefix", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}

		// --- When ---
		have := AddPrefixes("", data)

		// --- Then ---
		want := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		have := AddPrefixes("", nil)

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_ChangePrefix(t *testing.T) {
	t.Run("change", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"MY_KEY_3": "VAL_3",
		}

		// --- When ---
		have := ChangePrefix("MY", "YOUR", data)

		// --- Then ---
		want := map[string]string{
			"YOUR_KEY_1": "VAL_1",
			"KEY_2":      "VAL_2",
			"YOUR_KEY_3": "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("empty from", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}

		// --- When ---
		have := ChangePrefix("", "MY", data)

		// --- Then ---
		want := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("empty to", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{
			"MY_KEY_1": "VAL_1",
			"KEY_2":    "VAL_2",
			"KEY_3":    "VAL_3",
		}

		// --- When ---
		have := ChangePrefix("MY", "", data)

		// --- Then ---
		want := map[string]string{
			"KEY_1": "VAL_1",
			"KEY_2": "VAL_2",
			"KEY_3": "VAL_3",
		}
		assert.Equal(t, want, have)
		assert.NotSame(t, data, have)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		have := ChangePrefix("MY", "YOUR", nil)

		// --- Then ---
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

func Test_CheckDirExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "testdata")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("dir does not exist", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "/not/existing/dir")

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KEY_NAME: config key points to not existing directory"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldCode(t, "KEY_NAME", ECCfgKeyInvValue, err)
	})

	t.Run("dir is a file", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "testdata/equals.conf")

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KEY_NAME: config key points to not existing directory"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertFieldCode(t, "KEY_NAME", ECCfgKeyInvValue, err)
	})

	t.Run("empty", func(t *testing.T) {
		// --- When ---
		err := CheckDirExists("KEY_NAME", "")

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
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

	t.Run("relative path", func(t *testing.T) {
		// --- When ---
		err := CheckAbsPath("KEY_NAME", "abs/path")

		// --- Then ---
		wMsg := "KEY_NAME: config key must be an absolute path"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertHasField(t, "KEY_NAME", err)
	})

	t.Run("empty", func(t *testing.T) {
		// --- When ---
		err := CheckAbsPath("KEY_NAME", "")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "KEY_NAME", ErrKeyEmpty, err)
		assert.ErrorEqual(t, "KEY_NAME: config key not set", err)
	})
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
