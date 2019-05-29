// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

import (
	"fmt"

	"github.com/MaxHalford/eaopt"
)

type BoolSlice []bool

func (s BoolSlice) At(i int) interface{} {
	return s[i]
}

func (s BoolSlice) Set(i int, v interface{}) {
	s[i] = v.(bool)
}

func (s BoolSlice) Len() int {
	return len(s)
}

func (s BoolSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s BoolSlice) Slice(a, b int) eaopt.Slice {
	return s[a:b]
}

func (s BoolSlice) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return s[:k], s[k:]
}

func (s BoolSlice) Append(t eaopt.Slice) eaopt.Slice {
	return append(s, t.(BoolSlice)...)
}

func (s BoolSlice) Replace(t eaopt.Slice) {
	copy(s, t.(BoolSlice))
}

func (s BoolSlice) Copy() eaopt.Slice {
	t := make(BoolSlice, len(s))
	copy(t, s)
	return t
}

func (s BoolSlice) String() string {
	series, space := "", ""
	for i, value := range s {
		if value {
			series += fmt.Sprintf("%s%d", space, i+1)
			space = " "
		}
	}
	return series
}
