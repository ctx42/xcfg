// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfgtest

import (
	"strconv"
)

// TStringer is type implementing [fmt.Stringer] interface.
type TStringer int

func (v TStringer) String() string { return strconv.Itoa(int(v)) }
