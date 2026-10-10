// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

import (
	"maps"
	"strings"
)

// TODO(rz): Prefixes and key names should always go after maps, providers.
//   Now we have one convention in this file and other in adapters.

// GetOr returns the value of the specified key from the given provider. When
// the provider returns any error for the key, it returns the default value.
func GetOr(key string, cpv Provider, def string) string {
	if val, err := cpv.CfgKey(key); err == nil {
		return val
	}
	return def
}

// Prefix adds the prefix to the configuration key. If the prefix is a
// non-empty string, it is added to the key with the "_" separator.
func Prefix(prefix, key string) string {
	if prefix != "" {
		if key == "" {
			return prefix
		}
		key = prefix + "_" + key
	}
	return key
}

// StripPrefix strips the prefix and the "_" separator from the key. It
// reverses [Prefix] for every non-empty key; a key equal to the prefix is
// returned unchanged, although [Prefix] makes it from an empty key.
func StripPrefix(prefix, key string) string {
	if prefix != "" {
		key = strings.TrimPrefix(key, prefix+"_")
	}
	return key
}

// StripPrefixes strips the prefix from all the map keys. When a stripped key
// collides with a key that never had the prefix, the value of the prefixed key
// wins. It always returns a new map without editing the original one.
func StripPrefixes(prefix string, cfg map[string]string) map[string]string {
	if prefix == "" || len(cfg) == 0 {
		return maps.Clone(cfg)
	}
	ret := make(map[string]string, len(cfg))
	for key, val := range cfg {
		if StripPrefix(prefix, key) == key {
			ret[key] = val
		}
	}
	for key, val := range cfg {
		if stripped := StripPrefix(prefix, key); stripped != key {
			ret[stripped] = val
		}
	}
	return ret
}

// AddPrefixes adds the prefix to all the keys in the cfg map. It always
// returns a new map without editing the original one.
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
// When a changed key collides with an existing key, the value of the key that
// had the "from" prefix wins. It always returns a new map without editing the
// original one.
func ChangePrefix(from, to string, cfg map[string]string) map[string]string {
	if from == "" || len(cfg) == 0 {
		return maps.Clone(cfg)
	}
	ret := make(map[string]string, len(cfg))
	for key, val := range cfg {
		if StripPrefix(from, key) == key {
			ret[key] = val
		}
	}
	for key, val := range cfg {
		if stripped := StripPrefix(from, key); stripped != key {
			ret[Prefix(to, stripped)] = val
		}
	}
	return ret
}
