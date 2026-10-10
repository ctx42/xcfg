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
	_ Provider = Adaptor{}
	_ Adapter  = Adaptor{}
)

// Adaptor implements [Provider] and [Adapter] interfaces and provides
// a couple of configuration key adaptors.
type Adaptor struct {
	kind   string            // How to transform keys.
	from   string            // Key prefix to translate from.
	prefix string            // Key prefix to add or strip.
	tr     map[string]string // Translation map.
	cpv    Provider          // Provider to translate keys for.
}

func (adr Adaptor) CfgKey(key string) (string, error) {
	adapted := adr.CfgKeyAdapt(key)
	return adr.cpv.CfgKey(adapted)
}

func (adr Adaptor) CfgKeyAdapt(key string) string {
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
	if t, ok := adr.cpv.(Adapter); ok {
		return t.CfgKeyAdapt(key)
	}
	return key
}

// AddAdapter adapts keys by adding the prefix to every key you ask for.
func AddAdapter(cpv Provider, prefix string) Adaptor {
	return Adaptor{
		kind:   wAddPrefix,
		prefix: prefix,
		cpv:    cpv,
	}
}

// RemoveAdapter adapts keys by stripping the prefix from every key you ask for.
func RemoveAdapter(cpv Provider, prefix string) Adaptor {
	return Adaptor{
		kind:   wStripPrefix,
		prefix: prefix,
		cpv:    cpv,
	}
}

// MapAdapter adapts keys by translating them using given map.
func MapAdapter(cpv Provider, tr map[string]string) Adaptor {
	return Adaptor{
		kind: wTranslate,
		tr:   tr,
		cpv:  cpv,
	}
}

// ChangeAdapter adapts keys by changing key prefix "from" to "to".
func ChangeAdapter(cpv Provider, from, to string) Adaptor {
	return Adaptor{
		kind:   wPrefixTr,
		from:   from,
		prefix: to,
		cpv:    cpv,
	}
}
