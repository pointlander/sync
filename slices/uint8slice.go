// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type Uint8Slice []uint8

func (u Uint8Slice) At(i int) interface{} {
	return u[i]
}

func (u Uint8Slice) Set(i int, v interface{}) {
	u[i] = v.(uint8)
}

func (u Uint8Slice) Len() int {
	return len(u)
}

func (u Uint8Slice) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u Uint8Slice) Slice(a, b int) eaopt.Slice {
	return u[a:b]
}

func (u Uint8Slice) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return u[:k], u[k:]
}

func (u Uint8Slice) Append(t eaopt.Slice) eaopt.Slice {
	return append(u, t.(Uint8Slice)...)
}

func (u Uint8Slice) Replace(t eaopt.Slice) {
	copy(u, t.(Uint8Slice))
}

func (u Uint8Slice) Copy() eaopt.Slice {
	t := make(Uint8Slice, len(u))
	copy(t, u)
	return t
}

func (u Uint8Slice) String() string {
	series, space := "", ""
	for _, value := range u {
		series += fmt.Sprintf("%s%d", space, value)
		space = " "
	}
	return series
}
