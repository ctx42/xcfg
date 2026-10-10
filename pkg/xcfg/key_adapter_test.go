// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Adaptor(t *testing.T) {
	t.Run("chained", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "XX_B": "2"}
		cpv := Config(data)

		// --- When ---
		have := ChangeAdapter(cpv, "YY", "XX")
		have = AddAdapter(have, "YY")

		// --- Then ---
		value, err := have.CfgKey("B")
		assert.NoError(t, err)
		assert.Equal(t, "2", value)
	})

	t.Run("chained error", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "XX_B": "2"}
		cpv := Config(data)

		// --- When ---
		have := ChangeAdapter(cpv, "YY", "XX")
		have = AddAdapter(have, "YY")

		// --- Then ---
		value, err := have.CfgKey("C")
		xrrtest.AssertFieldIs(t, "XX_C", ErrKeyMissing, err)
		assert.ErrorEqual(t, "XX_C: config key missing", err)
		assert.Equal(t, "", value)
	})
}

func Test_AddAdapter(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "XX_B": "2"}
		cpv := Config(data)

		// --- When ---
		have := AddAdapter(cpv, "XX")

		// --- Then ---
		value, err := have.CfgKey("B")
		assert.NoError(t, err)
		assert.Equal(t, "2", value)
	})

	t.Run("not existing key", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "B": "2"}
		cpv := Config(data)

		// --- When ---
		have := AddAdapter(cpv, "XX")

		// --- Then ---
		value, err := have.CfgKey("NOT")
		xrrtest.AssertFieldIs(t, "XX_NOT", ErrKeyMissing, err)
		assert.ErrorEqual(t, "XX_NOT: config key missing", err)
		assert.Equal(t, "", value)
	})
}

func Test_RemoveAdapter(t *testing.T) {
	t.Run("key with prefix", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "B": "2"}
		cpv := Config(data)

		// --- When ---
		have := RemoveAdapter(cpv, "XX")

		// --- Then ---
		value, err := have.CfgKey("XX_A")
		assert.NoError(t, err)
		assert.Equal(t, "1", value)
	})

	t.Run("not existing key", func(t *testing.T) {
		// --- Given ---
		data := map[string]string{"A": "1", "B": "2"}
		cpv := Config(data)

		// --- When ---
		have := RemoveAdapter(cpv, "XX")

		// --- Then ---
		value, err := have.CfgKey("NOT")
		xrrtest.AssertFieldIs(t, "NOT", ErrKeyMissing, err)
		assert.ErrorEqual(t, "NOT: config key missing", err)
		assert.Equal(t, "", value)
	})
}

func Test_MapAdapter_tabular(t *testing.T) {
	tt := []struct {
		testN string

		cfg  map[string]string
		tr   map[string]string
		key  string
		want string
	}{
		{
			"key translated",
			map[string]string{"A": "1", "B": "2", "C": "3"},
			map[string]string{"B": "C"},
			"B",
			"3",
		},
		{
			"value of key not being translated",
			map[string]string{"A": "1", "B": "2", "C": "3"},
			map[string]string{"B": "C"},
			"C",
			"3",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cpv := MapAdapter(Config(tc.cfg), tc.tr)

			// --- When ---
			have, err := cpv.CfgKey(tc.key)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_MapAdapter(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		// --- Given ---
		cfg := map[string]string{"A": "1", "B": "2", "C": "3"}
		cpv := MapAdapter(Config(cfg), nil)

		// --- When ---
		have, err := cpv.CfgKey("A")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "1", have)
	})

	t.Run("not existing key", func(t *testing.T) {
		// --- Given ---
		cfg := map[string]string{"A": "1", "B": "2", "C": "3"}
		cpv := MapAdapter(Config(cfg), nil)

		// --- When ---
		have, err := cpv.CfgKey("D")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "D", ErrKeyMissing, err)
		assert.Equal(t, "", have)
	})

	t.Run("not existing translated key", func(t *testing.T) {
		// --- Given ---
		cfg := map[string]string{"A": "1", "B": "2", "C": "3"}
		cpv := MapAdapter(Config(cfg), map[string]string{"A": "D"})

		// --- When ---
		have, err := cpv.CfgKey("A")

		// --- Then ---
		xrrtest.AssertFieldIs(t, "D", ErrKeyMissing, err)
		assert.Equal(t, "", have)
	})
}

func Test_ChangeAdapter_tabular(t *testing.T) {
	tt := []struct {
		testN string

		cfg  map[string]string
		from string
		to   string
		key  string
		want string
	}{
		{
			"prefix substituted",
			map[string]string{"XX_A": "1", "A": "2"},
			"YY",
			"XX",
			"XX_A",
			"1",
		},
		{
			"key is prefixed only if from prefix is present",
			map[string]string{"XX_A": "1", "A": "2"},
			"YY",
			"XX",
			"A",
			"2",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			cpv := ChangeAdapter(Config(tc.cfg), tc.from, tc.to)

			// --- When ---
			have, err := cpv.CfgKey(tc.key)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}
