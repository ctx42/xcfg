// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg

// Wrap kinds.
const (
	wAddPrefix   = "wrap-add-prefix"
	wStripPrefix = "wrap-strip-prefix"
	wTranslate   = "wrap-tr"
	wPrefixTr    = "wrap-prefix-tr"
)

// Compile time checks.
var (
	_ Provider = KeyAdapter{}
	_ Adapter  = KeyAdapter{}
)

// KeyAdapter is a [Provider] that rewrites the key before the lookup in the
// provider it wraps. Create it with [AddAdapter], [RemoveAdapter],
// [MapAdapter], or [ChangeAdapter].
type KeyAdapter struct {
	kind   string            // How to transform keys.
	from   string            // Key prefix to translate from.
	prefix string            // Key prefix to add or strip.
	tr     map[string]string // Translation map.
	cpv    Provider          // Provider to translate keys for.
}

// CfgKey returns [ErrKeyMissing] for every key when the [KeyAdapter] is the
// zero value, which wraps no provider.
func (adr KeyAdapter) CfgKey(key string) (string, error) {
	key = adr.adapt(key)
	if adr.cpv == nil {
		return "", NewFieldError(key, ErrKeyMissing)
	}
	return adr.cpv.CfgKey(key)
}

// CfgKeyAdapt returns the key after the rewrite of this [KeyAdapter] and of
// every [Adapter] it wraps.
func (adr KeyAdapter) CfgKeyAdapt(key string) string {
	key = adr.adapt(key)
	if t, ok := adr.cpv.(Adapter); ok {
		return t.CfgKeyAdapt(key)
	}
	return key
}

// adapt applies this adapter's key transformation only, without the
// transformations of the adapters it wraps.
func (adr KeyAdapter) adapt(key string) string {
	switch adr.kind {
	case wAddPrefix:
		key = Prefix(adr.prefix, key)
	case wStripPrefix:
		key = StripPrefix(adr.prefix, key)
	case wTranslate:
		if val, ok := adr.tr[key]; ok {
			key = val
		}
	case wPrefixTr:
		keyS := StripPrefix(adr.from, key)
		// Add prefix only if the "from" prefix was removed.
		if keyS != key {
			key = Prefix(adr.prefix, keyS)
		}
	}
	return key
}

// AddAdapter adapts keys by adding the prefix to every key you ask for.
func AddAdapter(cpv Provider, prefix string) KeyAdapter {
	return KeyAdapter{
		kind:   wAddPrefix,
		prefix: prefix,
		cpv:    cpv,
	}
}

// RemoveAdapter adapts keys by stripping the prefix from every key you ask for.
func RemoveAdapter(cpv Provider, prefix string) KeyAdapter {
	return KeyAdapter{
		kind:   wStripPrefix,
		prefix: prefix,
		cpv:    cpv,
	}
}

// MapAdapter adapts keys by translating them using given map.
func MapAdapter(cpv Provider, tr map[string]string) KeyAdapter {
	return KeyAdapter{
		kind: wTranslate,
		tr:   tr,
		cpv:  cpv,
	}
}

// ChangeAdapter adapts keys by changing key prefix "from" to "to".
func ChangeAdapter(cpv Provider, from, to string) KeyAdapter {
	return KeyAdapter{
		kind:   wPrefixTr,
		from:   from,
		prefix: to,
		cpv:    cpv,
	}
}
