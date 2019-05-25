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
	Places = 6
	// FixedOne is the number 1
	FixedOne = 1 << Places
	// FixedHalf is the number .5
	FixedHalf = 1 << (Places - 1)
)

// Fixed is a fixed point number
type Fixed int32

var factors [31]float64

func init() {
	factor := 2
	for i := uint(0); i < Places; i++ {
		factors[Places-1-i] = 1 / float64(factor)
		factor *= 2
	}
	factor = 1
	for i := uint(Places); i < 31; i++ {
		factors[i] = float64(factor)
		factor *= 2
	}
}

// Float64 converts the fixed number to a float64
func (f Fixed) Float64() float64 {
	value, sign := .0, false
	if f < 0 {
		f, sign = -f, true
	}
	for _, v := range factors {
		if f&1 == 1 {
			value += v
		}
		f >>= 1
	}
	if sign {
		value = -value
	}
	return value
}

// String converts the fixed point number to a string
func (f Fixed) String() string {
	return fmt.Sprintf("%f", f.Float64())
}

const (
	// Float64ExponentMask is the float64 exponent mask
	Float64ExponentMask = 1<<11 - 1
	// Float64FractionMask is the float64 fraction mask
	Float64FractionMask = 1<<52 - 1
	// Float64Bias is the float64 exponent bias
	Float64Bias = 1<<10 - 1
	// Fixed32Mask is the fixed mask
	Fixed32Mask = 1<<32 - 1
)

// FixedFromFloat64 creates a fixed point number from a float64
func FixedFromFloat64(a float64) Fixed {
	bits := math.Float64bits(a)
	sign := bits >> 63
	exponent := int((bits>>52)&Float64ExponentMask) - Float64Bias
	fraction := bits&Float64FractionMask | 1<<53
	if exponent > (32 - Places) {
		panic("exponent is too larger")
	} else if exponent < -Places {
		panic("exponent is too small")
	}
	shift := uint(53 - exponent - Places)
	fixed := Fixed((fraction >> shift) & Fixed32Mask)
	if sign != 0 {
		fixed = -fixed
	}
	return fixed
}

// Mul multiplys to fixed point nuimbers
func (f Fixed) Mul(b Fixed) Fixed {
	return Fixed((int64(f)*int64(b) + FixedHalf) >> Places)
}
