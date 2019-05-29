// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type Uint8 []uint8

func (u Uint8) At(i int) interface{} {
	return u[i]
}

func (u Uint8) Set(i int, v interface{}) {
	u[i] = v.(uint8)
}

func (u Uint8) Len() int {
	return len(u)
}

func (u Uint8) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u Uint8) Slice(a, b int) eaopt.Slice {
	return u[a:b]
}

func (u Uint8) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return u[:k], u[k:]
}

func (u Uint8) Append(t eaopt.Slice) eaopt.Slice {
	return append(u, t.(Uint8)...)
}

func (u Uint8) Replace(t eaopt.Slice) {
	copy(u, t.(Uint8))
}

func (u Uint8) Copy() eaopt.Slice {
	t := make(Uint8, len(u))
	copy(t, u)
	return t
}

func (u Uint8) String() string {
	series, space := "", ""
	for _, value := range u {
		series += fmt.Sprintf("%s%d", space, value)
		space = " "
	}
	return series
}
