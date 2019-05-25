// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fixed

import "testing"

func TestFixedFloat32(t *testing.T) {
	vectors := [...]float32{4, -4, 2, -2, .5, -.5, .25, -.25}
	for _, v := range vectors {
		a := FixedFloat32(v)
		if b := a.Float32(); b != v {
			t.Errorf("%f != %f", v, b)
		}
	}
}

func TestFixedFloat64(t *testing.T) {
	vectors := [...]float64{4, -4, 2, -2, .5, -.5, .25, -.25}
	for _, v := range vectors {
		a := FixedFloat64(v)
		if b := a.Float64(); b != v {
			t.Errorf("%f != %f", v, b)
		}
	}
}

func TestFixed_Mul(t *testing.T) {
	a := Fixed(FixedHalf)
	b := a.Mul(a)
	if b.Float64() != .25 {
		t.Fatalf("%s != .25", b)
	}
}

var z Fixed

func BenchmarkFixedMul(t *testing.B) {
	x, y := Fixed(FixedOne), Fixed(FixedHalf)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		z = x.Mul(y)
		x++
		y++
	}
}

var z32 float32

func BenchmarkMulFloat32(t *testing.B) {
	x, y := float32(1.0), float32(.5)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		z32 = x * y
		x += 1 / 64
		y += 1 / 64
	}
}

var z64 float64

func BenchmarkMulFloat64(t *testing.B) {
	x, y := 1.0, .5
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		z64 = x * y
		x += 1 / 64
		y += 1 / 64
	}
}
