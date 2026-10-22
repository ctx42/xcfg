// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"syscall"
	"unicode"

	"github.com/ctx42/dotenv/pkg/dotenv"
)

// Open reads the whole file at pth and returns a reader over its content.
func Open(pth string) (*bytes.Reader, error) {
	buf, err := os.ReadFile(pth) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// Merge merges keys from src to destination and returns destination.
func Merge(dst map[string]string, src ...map[string]string) map[string]string {
	for _, s := range src {
		for k, v := range s {
			if dst == nil {
				dst = make(map[string]string, len(s))
			}
			dst[k] = v
		}
	}
	return dst
}

// DirExists returns true if pth points to an existing directory. A path that
// does not exist, or that runs through a regular file, is not an error.
func DirExists(pth string) (bool, error) {
	fi, err := os.Stat(pth)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
			return false, nil
		}
		return false, err
	}

	mode := fi.Mode()
	return mode.IsDir(), nil
}

// doubleQuoteSpecialChars represents list of characters to escape.
const doubleQuoteSpecialChars = "\\\n\r\"!$`"

// doubleQuoteEscape escapes characters on the doubleQuoteSpecialChars list.
func doubleQuoteEscape(lin string) string {
	for _, c := range doubleQuoteSpecialChars {
		toReplace := "\\" + string(c)
		if c == '\n' {
			toReplace = `\n`
		}
		if c == '\r' {
			toReplace = `\r`
		}
		lin = strings.ReplaceAll(lin, string(c), toReplace)
	}
	return lin
}

// isDotenvKey reports whether [New] reads the key back as written: it is not
// empty and holds no "=", ":", "#", or whitespace.
func isDotenvKey(key string) bool {
	if key == "" || strings.ContainsAny(key, "=:#") {
		return false
	}
	return strings.IndexFunc(key, unicode.IsSpace) == -1
}

// maybeDecode decodes Base64 encoded config, if needed, and returns a reader
// to the config. It accepts the encoded data with "=" padding and surrounding
// whitespace.
func maybeDecode(r io.Reader) (io.Reader, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if enc, ok := bytes.CutPrefix(src, []byte(prefix64)); ok {
		enc = bytes.TrimRight(bytes.TrimSpace(enc), "=")
		dec := base64.RawURLEncoding
		src = make([]byte, dec.DecodedLen(len(enc)))
		var n int
		if n, err = dec.Decode(src, enc); err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
		src = src[:n]
	}
	return bytes.NewReader(src), nil
}

// EnvSplit parses [os.Environ] results and returns it as key value map.
func EnvSplit(env []string) map[string]string { return dotenv.Split(env) }
