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

func Test_ParseTime_tabular_errors(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		tim     string
		wantMsg string
	}{
		{
			"empty",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"invalid format",
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
			assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			xrrtest.AssertHasField(t, tc.key, err)
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

func Test_ParseDur_tabular_errors(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		dur     string
		wantMsg string
	}{
		{
			"1",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"2",
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
			assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			xrrtest.AssertHasField(t, tc.key, err)
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

	t.Run("error - empty string", func(t *testing.T) {
		// --- When ---
		have, err := ParseInteger("KEY_NAME", "")

		// --- Then ---
		assert.ErrorIs(t, ErrKeyEmpty, err)
		assert.Equal(t, 0, have)
	})
}

func Test_ParseInteger_tabular_errors(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		dur     string
		wantMsg string
	}{
		{
			"1",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"2",
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
			have, err := ParseInteger(tc.key, tc.dur)

			// --- Then ---
			assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			xrrtest.AssertHasField(t, tc.key, err)
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

	t.Run("error - empty string", func(t *testing.T) {
		// --- When ---
		have, err := ParseTimezone("KEY_NAME", "")

		// --- Then ---
		assert.ErrorIs(t, ErrKeyEmpty, err)
		assert.Nil(t, have)
	})
}

func Test_ParseTimezone_tabular_errors(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		tim     string
		wantMsg string
	}{
		{
			"1",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"2",
			"KEY_NAME",
			"Unknown/Timezone",
			"" +
				"KEY_NAME: invalid config key timezone value: " +
				"unknown time zone Unknown/Timezone (ECCfgKeyInvValue)",
		},
		{
			"3",
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
			assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			xrrtest.AssertHasField(t, tc.key, err)
			assert.Nil(t, have)
		})
	}
}

func Test_ParseBool_tabular_success(t *testing.T) {
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

func Test_ParseBool_tabular_errors(t *testing.T) {
	tt := []struct {
		testN string

		key     string
		dur     string
		wantMsg string
	}{
		{
			"1",
			"KEY_NAME",
			"",
			"KEY_NAME: config key not set (ECCfgKeyEmpty)",
		},
		{
			"2",
			"KEY_NAME",
			"abc",
			"" +
				"KEY_NAME: invalid config key boolean format: " +
				"strconv.ParseBool: parsing \"abc\": invalid syntax " +
				"(ECCfgKeyInvFormat)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := ParseBool(tc.key, tc.dur)

			// --- Then ---
			assert.SameType(t, &FieldErrors{}, err)
			xrrtest.AssertEqual(t, tc.wantMsg, err)
			xrrtest.AssertHasField(t, tc.key, err)
			assert.Zero(t, have)
		})
	}
}
