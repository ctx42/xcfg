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
		err = NewError("invalid config key RFC3339 format", ECCfgKeyInvFormat)
		return time.Time{}, NewFieldError(key, err)
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
		err = NewError("invalid config key duration format", ECCfgKeyInvFormat)
		return 0, NewFieldError(key, err)
	}
	return t, nil
}

// ParseInteger parses a configuration key as integer.
func ParseInteger(key, val string) (int, error) {
	if val == "" {
		return 0, NewFieldError(key, ErrKeyEmpty)
	}

	i, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		err = NewError("invalid config key integer format", ECCfgKeyInvFormat)
		return 0, NewFieldError(key, err)
	}
	return int(i), nil
}

// ParseTimezone parses a configuration key as timezone.
func ParseTimezone(key, val string) (*time.Location, error) {
	if val == "" {
		return nil, NewFieldError(key, ErrKeyEmpty)
	}

	loc, err := time.LoadLocation(val)
	if err != nil {
		err = NewError("invalid config key timezone value", ECCfgKeyInvValue)
		return nil, NewFieldError(key, err)
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
		err = NewError("invalid config key boolean format", ECCfgKeyInvFormat)
		return false, NewFieldError(key, err)
	}
	return b, nil
}
