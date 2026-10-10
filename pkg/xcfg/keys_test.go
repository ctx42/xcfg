// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

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
			cfg := NewFromEnv(tc.env)

			// --- When ---
			have := GetOr(tc.key, cfg, tc.def)

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
			// --- When ---
			have := Prefix(tc.prefix, tc.key)

			// --- Then ---
			assert.Equal(t, tc.want, have)
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
			// --- When ---
			have := StripPrefix(tc.prefix, tc.key)

			// --- Then ---
			assert.Equal(t, tc.want, have)
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

	t.Run("prefixed key wins on collision", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"MY_KEY": "VAL_1", "KEY": "VAL_2"}

		// --- When ---
		have := StripPrefixes("MY", data)

		// --- Then ---
		assert.Equal(t, map[string]string{"KEY": "VAL_1"}, have)
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

	t.Run("key starting with from without separator", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"MYSQL_HOST": "VAL_1", "MY": "VAL_2"}

		// --- When ---
		have := ChangePrefix("MY", "YOUR", data)

		// --- Then ---
		want := map[string]string{"MYSQL_HOST": "VAL_1", "MY": "VAL_2"}
		assert.Equal(t, want, have)
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

	t.Run("prefixed key wins on collision", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"MY_KEY": "VAL_1", "YOUR_KEY": "VAL_2"}

		// --- When ---
		have := ChangePrefix("MY", "YOUR", data)

		// --- Then ---
		assert.Equal(t, map[string]string{"YOUR_KEY": "VAL_1"}, have)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		have := ChangePrefix("MY", "YOUR", nil)

		// --- Then ---
		assert.Nil(t, have)
	})
}
