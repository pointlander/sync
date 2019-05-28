// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
)

type FixedSlice []fixed.Fixed

func (f FixedSlice) At(i int) interface{} {
	return f[i]
}

func (f FixedSlice) Set(i int, v interface{}) {
	f[i] = v.(fixed.Fixed)
}

func (f FixedSlice) Len() int {
	return len(f)
}

func (f FixedSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FixedSlice) Slice(a, b int) eaopt.Slice {
	return f[a:b]
}

func (f FixedSlice) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return f[:k], f[k:]
}

func (f FixedSlice) Append(t eaopt.Slice) eaopt.Slice {
	return append(f, t.(FixedSlice)...)
}

func (f FixedSlice) Replace(t eaopt.Slice) {
	copy(f, t.(FixedSlice))
}

func (f FixedSlice) Copy() eaopt.Slice {
	t := make(FixedSlice, len(f))
	copy(t, f)
	return t
}

func (f FixedSlice) String() string {
	series, space := "", ""
	for _, value := range f {
		series += fmt.Sprintf("%s%f", space, value)
		space = " "
	}
	return series
}
