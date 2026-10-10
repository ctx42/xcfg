// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"bytes"
	"encoding/base64"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctx42/dotenv/pkg/dotenv"
	"github.com/ctx42/xrr/pkg/xrr"
)

// TODO(rz): Prefixes and key names should always go after maps, providers.
//   Now we have one convention in this file and other in adapters.

// Open helper function for reading a file to buffer and returning a reader.
func Open(pth string) (io.Reader, error) {
	buf, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// GetOr returns the value of the specified key from the given provider. If the
// key is not found in the provider, it returns the provided default value.
func GetOr(key string, cpv Provider, def string) string {
	if val, err := cpv.CfgKey(key); err == nil {
		return val
	}
	return def
}

// Prefix adds the prefix to the configuration key. If the prefix is non-empty
// string, it will be added with "_" separator to the key.
func Prefix(prefix, key string) string {
	if prefix != "" {
		if key == "" {
			return prefix
		}
		key = prefix + "_" + key
	}
	return key
}

// StripPrefix strips the prefix from the key. It is the reverse to what
// [Prefix] does.
func StripPrefix(prefix, key string) string {
	if prefix != "" {
		key = strings.TrimPrefix(key, prefix+"_")
	}
	return key
}

// StripPrefixes strips prefix from all the map keys. It does not edit the
// original map. It always returns a new map without editing the original one.
func StripPrefixes(prefix string, cfg map[string]string) map[string]string {
	if prefix == "" || len(cfg) == 0 {
		return maps.Clone(cfg)
	}
	ret := make(map[string]string, len(cfg))
	for key, val := range cfg {
		ret[StripPrefix(prefix, key)] = val
	}
	return ret
}

// AddPrefixes adds prefix to all the keys in the cfg map. It always returns a
// new map without editing the original one.
func AddPrefixes(prefix string, cfg map[string]string) map[string]string {
	if prefix == "" || len(cfg) == 0 {
		return maps.Clone(cfg)
	}
	ret := make(map[string]string, len(cfg))
	for key, val := range cfg {
		ret[Prefix(prefix, key)] = val
	}
	return ret
}

// ChangePrefix changes all key names which have prefix "from" to prefix "to".
// It always returns a new map without editing the original one.
func ChangePrefix(from, to string, cfg map[string]string) map[string]string {
	if from == "" || len(cfg) == 0 {
		return maps.Clone(cfg)
	}
	ret := make(map[string]string, len(cfg))
	for key, val := range cfg {
		if strings.HasPrefix(key, from) {
			stripped := StripPrefix(from, key)
			ret[Prefix(to, stripped)] = val
		} else {
			ret[key] = val
		}
	}
	return ret
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

// DirExists returns true if pth points to an existing directory.
func DirExists(pth string) (bool, error) {
	fi, err := os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	mode := fi.Mode()
	return mode.IsDir(), nil
}

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
		msg := "config key points to not existing directory"
		err = NewError(msg, ECCfgKeyInvValue)
		return NewFieldError(key, err)
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
		msg := "config key must be an absolute path"
		err := NewError(msg, ECCfgKeyInvValue)
		return NewFieldError(key, err)
	}
	return nil
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

// maybeDecode decodes Base64 encoded config, if needed, and returns a reader
// to the config.
func maybeDecode(r io.Reader) (io.Reader, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if enc, ok := bytes.CutPrefix(src, []byte(prefix64)); ok {
		dec := base64.RawURLEncoding
		src = make([]byte, dec.DecodedLen(len(enc)))
		var n int
		if n, err = dec.Decode(src, enc); err != nil {
			return nil, err
		}
		src = src[:n]
	}
	return bytes.NewReader(src), nil
}

// EnvSplit parses [os.Environ] results and returns it as key value map.
func EnvSplit(env []string) map[string]string { return dotenv.Split(env) }
