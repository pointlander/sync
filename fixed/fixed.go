// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fixed

import (
	"fmt"
	"math"
)

const (
	// Places is the number of fixed places
	Places = 16
	// FixedOne is the number 1
	FixedOne = 1 << Places
	// FixedHalf is the number .5
	FixedHalf = 1 << (Places - 1)
)

// Fixed is a fixed point number
type Fixed int32

// FixedFromFloat32 creates a fixed point number from a float32
func FixedFloat32(a float32) Fixed {
	round := float32(.5)
	if a < 0 {
		round = -0.5
	}
	b := a*FixedOne + round
	if b > math.MaxInt32 {
		panic(fmt.Errorf("float is too big: %f", a))
	} else if b < math.MinInt32 {
		panic(fmt.Errorf("float is too small %f", a))
	}
	return Fixed(b)
}

// FixedFromFloat64 creates a fixed point number from a float64
func FixedFloat64(a float64) Fixed {
	round := .5
	if a < 0 {
		round = -0.5
	}
	b := a*FixedOne + round
	if b > math.MaxInt32 {
		panic(fmt.Errorf("float is too big: %f", a))
	} else if b < math.MinInt32 {
		panic(fmt.Errorf("float is too small %f", a))
	}
	return Fixed(b)
}

// Float32 converts the fixed number to a float64
func (f Fixed) Float32() float32 {
	return float32(f) / FixedOne
}

// Float64 converts the fixed number to a float64
func (f Fixed) Float64() float64 {
	return float64(f) / FixedOne
}

// String converts the fixed point number to a string
func (f Fixed) String() string {
	return fmt.Sprintf("%f", f.Float32())
}

// Abs returns the absolute value
func (f Fixed) Abs() Fixed {
	if f < 0 {
		return -f
	}
	return f
}

// Mul multiplys to fixed point nuimbers
func (f Fixed) Mul(b Fixed) Fixed {
	return Fixed((int64(f)*int64(b) + FixedHalf) >> Places)
}
