// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"errors"
	"path/filepath"

	"github.com/ctx42/xrr/pkg/xrr"
)

// CheckDirExists returns an error if the configuration key is not a directory
// path, or it's an empty string.
func CheckDirExists(key, pth string) error {
	if pth == "" {
		return NewFieldError(key, ErrKeyEmpty)
	}
	ok, err := DirExists(pth)
	if err != nil {
		err = NewError("config", ECCfgError, xrr.WithCause(err))
		return NewFieldError(key, err)
	}
	if !ok {
		err = errors.New("directory does not exist")
		return keyError(key, ErrKeyInvalid, err)
	}
	return nil
}

// CheckAbsPath returns an error if the configuration key is not an absolute
// path, or it's an empty string.
func CheckAbsPath(key, pth string) error {
	if pth == "" {
		return NewFieldError(key, ErrKeyEmpty)
	}
	if !filepath.IsAbs(pth) {
		return keyError(key, ErrKeyInvalid, errors.New("not an absolute path"))
	}
	return nil
}
