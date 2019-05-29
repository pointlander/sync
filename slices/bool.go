// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type Bool []bool

func (s Bool) At(i int) interface{} {
	return s[i]
}

func (s Bool) Set(i int, v interface{}) {
	s[i] = v.(bool)
}

func (s Bool) Len() int {
	return len(s)
}

func (s Bool) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Bool) Slice(a, b int) eaopt.Slice {
	return s[a:b]
}

func (s Bool) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return s[:k], s[k:]
}

func (s Bool) Append(t eaopt.Slice) eaopt.Slice {
	return append(s, t.(Bool)...)
}

func (s Bool) Replace(t eaopt.Slice) {
	copy(s, t.(Bool))
}

func (s Bool) Copy() eaopt.Slice {
	t := make(Bool, len(s))
	copy(t, s)
	return t
}

func (s Bool) String() string {
	series, space := "", ""
	for i, value := range s {
		if value {
			series += fmt.Sprintf("%s%d", space, i+1)
			space = " "
		}
	}
	return series
}
