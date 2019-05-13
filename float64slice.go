// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type Float64Slice []float64

func (f Float64Slice) At(i int) interface{} {
	return f[i]
}

func (f Float64Slice) Set(i int, v interface{}) {
	f[i] = v.(float64)
}

func (f Float64Slice) Len() int {
	return len(f)
}

func (f Float64Slice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f Float64Slice) Slice(a, b int) eaopt.Slice {
	return f[a:b]
}

func (f Float64Slice) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return f[:k], f[k:]
}

func (f Float64Slice) Append(t eaopt.Slice) eaopt.Slice {
	return append(f, t.(Float64Slice)...)
}

func (f Float64Slice) Replace(t eaopt.Slice) {
	copy(f, t.(Float64Slice))
}

func (f Float64Slice) Copy() eaopt.Slice {
	t := make(Float64Slice, len(f))
	copy(t, f)
	return t
}

func (f Float64Slice) String() string {
	series, space := "", ""
	for _, value := range f {
		series += fmt.Sprintf("%s%f", space, value)
		space = " "
	}
	return series
}
