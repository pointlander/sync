// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
)

type Fixed []fixed.Fixed

func (f Fixed) At(i int) interface{} {
	return f[i]
}

func (f Fixed) Set(i int, v interface{}) {
	f[i] = v.(fixed.Fixed)
}

func (f Fixed) Len() int {
	return len(f)
}

func (f Fixed) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f Fixed) Slice(a, b int) eaopt.Slice {
	return f[a:b]
}

func (f Fixed) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return f[:k], f[k:]
}

func (f Fixed) Append(t eaopt.Slice) eaopt.Slice {
	return append(f, t.(Fixed)...)
}

func (f Fixed) Replace(t eaopt.Slice) {
	copy(f, t.(Fixed))
}

func (f Fixed) Copy() eaopt.Slice {
	t := make(Fixed, len(f))
	copy(t, f)
	return t
}

func (f Fixed) String() string {
	series, space := "", ""
	for _, value := range f {
		series += fmt.Sprintf("%s%f", space, value)
		space = " "
	}
	return series
}
