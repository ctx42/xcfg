// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"strconv"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_ParseTime(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tim := "2021-09-10T13:23:24.000000123+02:00"

		// --- When ---
		have, err := ParseTime("KEY_NAME", tim)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, tim, have.Format(time.RFC3339Nano))
	})
}

func Test_ParseTime_tabular(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		tim     string
		wantMsg string
	}{
		{
			"error - empty",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"error - invalid format",
			"KEY_NAME",
			"2001-01-02",
			"" +
				"KEY_NAME: invalid config key RFC3339 format: " +
				"parsing time \"2001-01-02\" as " +
				"\"2006-01-02T15:04:05.999999999Z07:00\": " +
				"cannot parse \"\" as \"T\" (ECCfgKeyInvFormat)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ParseTime(tc.key, tc.tim)

			// --- Then ---
			_, _ = assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			_, _ = xrrtest.AssertHasField(t, tc.key, err)
			assert.Zero(t, have)
		})
	}
}

func Test_ParseDur(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		want := 6*time.Hour + 2*time.Second

		// --- When ---
		have, err := ParseDur("KEY_NAME", "6h2s")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}

func Test_ParseDur_tabular(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		dur     string
		wantMsg string
	}{
		{
			"error - empty",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"error - invalid format",
			"KEY_NAME",
			"abc",
			"" +
				"KEY_NAME: invalid config key duration format: " +
				"time: invalid duration \"abc\" (ECCfgKeyInvFormat)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ParseDur(tc.key, tc.dur)

			// --- Then ---
			_, _ = assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			_, _ = xrrtest.AssertHasField(t, tc.key, err)
			assert.Zero(t, have)
		})
	}
}

func Test_ParseInteger(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		have, err := ParseInteger("KEY_NAME", "123")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 123, have)
	})

	t.Run("beyond int32", func(t *testing.T) {
		// --- When ---
		have, err := ParseInteger("KEY_NAME", "3000000000")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 3000000000, have)
	})

	t.Run("error - cause kept", func(t *testing.T) {
		// --- When ---
		have, err := ParseInteger("KEY_NAME", "abc")

		// --- Then ---
		assert.ErrorIs(t, strconv.ErrSyntax, err)
		assert.Equal(t, 0, have)
	})
}

func Test_ParseInteger_tabular(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		val     string
		wantMsg string
	}{
		{
			"error - empty",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"error - invalid format",
			"KEY_NAME",
			"abc",
			"" +
				"KEY_NAME: invalid config key integer format: " +
				"strconv.Atoi: parsing \"abc\": invalid syntax " +
				"(ECCfgKeyInvFormat)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ParseInteger(tc.key, tc.val)

			// --- Then ---
			_, _ = assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			_, _ = xrrtest.AssertHasField(t, tc.key, err)
			assert.Zero(t, have)
		})
	}
}

func Test_ParseTimezone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		have, err := ParseTimezone("KEY_NAME", "Europe/Warsaw")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "Europe/Warsaw", have.String())
	})
}

func Test_ParseTimezone_tabular(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		tim     string
		wantMsg string
	}{
		{
			"error - empty",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"error - unknown zone",
			"KEY_NAME",
			"Unknown/Timezone",
			"" +
				"KEY_NAME: invalid config key timezone value: " +
				"unknown time zone Unknown/Timezone (ECCfgKeyInvValue)",
		},
		{
			"error - invalid name",
			"KEY_NAME",
			"123 abc",
			"" +
				"KEY_NAME: invalid config key timezone value: " +
				"unknown time zone 123 abc (ECCfgKeyInvValue)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ParseTimezone(tc.key, tc.tim)

			// --- Then ---
			_, _ = assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			_, _ = xrrtest.AssertHasField(t, tc.key, err)
			assert.Nil(t, have)
		})
	}
}

func Test_ParseBool(t *testing.T) {
	t.Run("error - empty", func(t *testing.T) {
		// --- When ---
		have, err := ParseBool("KEY_NAME", "")

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KEY_NAME: config key not set (ECCfgKeyEmpty)"
		xrrtest.AssertEqual(t, wMsg, err)
		_, _ = xrrtest.AssertHasField(t, "KEY_NAME", err)
		assert.False(t, have)
	})

	t.Run("error - invalid format", func(t *testing.T) {
		// --- When ---
		have, err := ParseBool("KEY_NAME", "abc")

		// --- Then ---
		_, _ = assert.SameType(t, &FieldErrors{}, err)
		wMsg := "" +
			"KEY_NAME: invalid config key boolean format: " +
			"strconv.ParseBool: parsing \"abc\": invalid syntax " +
			"(ECCfgKeyInvFormat)"
		xrrtest.AssertEqual(t, wMsg, err)
		_, _ = xrrtest.AssertHasField(t, "KEY_NAME", err)
		assert.False(t, have)
	})
}

func Test_ParseBool_tabular(t *testing.T) {
	tt := []struct {
		val  string
		want bool
	}{
		{"1", true},
		{"t", true},
		{"T", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"0", false},
		{"f", false},
		{"F", false},
		{"FALSE", false},
		{"false", false},
		{"False", false},
		{"no", false},
		{"No", false},
		{"NO", false},
	}

	for _, tc := range tt {
		t.Run(tc.val, func(t *testing.T) {
			// --- When ---
			have, err := ParseBool("KEY_NAME", tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}
