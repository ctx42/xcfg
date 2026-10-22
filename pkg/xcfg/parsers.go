// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"strconv"
	"time"
)

// ParseTime parses a configuration key as RFC3339 time.
func ParseTime(key, val string) (time.Time, error) {
	if val == "" {
		return time.Time{}, NewFieldError(key, ErrKeyEmpty)
	}

	tim, err := time.Parse(time.RFC3339Nano, val)
	if err != nil {
		return time.Time{}, keyError(key, ErrKeyInvFormat, err)
	}
	return tim, nil
}

// ParseDur parses a configuration key as a time.Duration.
func ParseDur(key, val string) (time.Duration, error) {
	if val == "" {
		return 0, NewFieldError(key, ErrKeyEmpty)
	}

	var t time.Duration
	var err error
	if t, err = time.ParseDuration(val); err != nil {
		return 0, keyError(key, ErrKeyInvFormat, err)
	}
	return t, nil
}

// ParseInteger parses a configuration key as integer.
func ParseInteger(key, val string) (int, error) {
	if val == "" {
		return 0, NewFieldError(key, ErrKeyEmpty)
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, keyError(key, ErrKeyInvFormat, err)
	}
	return i, nil
}

// ParseTimezone parses a configuration key as timezone.
func ParseTimezone(key, val string) (*time.Location, error) {
	if val == "" {
		return nil, NewFieldError(key, ErrKeyEmpty)
	}

	loc, err := time.LoadLocation(val)
	if err != nil {
		return nil, keyError(key, ErrKeyInvalid, err)
	}
	return loc, nil
}

// ParseBool parses a configuration key as boolean.
// It accepts 1, t, T, TRUE, true, True, yes, Yes, YES, 0, f, F, FALSE, false,
// False, no, No, NO.
func ParseBool(key, val string) (bool, error) {
	if val == "" {
		return false, NewFieldError(key, ErrKeyEmpty)
	}

	if val == "yes" || val == "Yes" || val == "YES" {
		return true, nil
	}

	if val == "no" || val == "No" || val == "NO" {
		return false, nil
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, keyError(key, ErrKeyInvFormat, err)
	}
	return b, nil
}
