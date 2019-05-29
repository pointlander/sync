// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type Float64 []float64

func (f Float64) At(i int) interface{} {
	return f[i]
}

func (f Float64) Set(i int, v interface{}) {
	f[i] = v.(float64)
}

func (f Float64) Len() int {
	return len(f)
}

func (f Float64) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f Float64) Slice(a, b int) eaopt.Slice {
	return f[a:b]
}

func (f Float64) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return f[:k], f[k:]
}

func (f Float64) Append(t eaopt.Slice) eaopt.Slice {
	return append(f, t.(Float64)...)
}

func (f Float64) Replace(t eaopt.Slice) {
	copy(f, t.(Float64))
}

func (f Float64) Copy() eaopt.Slice {
	t := make(Float64, len(f))
	copy(t, f)
	return t
}

func (f Float64) String() string {
	series, space := "", ""
	for _, value := range f {
		series += fmt.Sprintf("%s%f", space, value)
		space = " "
	}
	return series
}
