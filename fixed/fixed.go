package fixed

import "fmt"

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

var factors [32]float64

func init() {
	factor := 2
	for i := uint(0); i < Places; i++ {
		factors[Places-1-i] = 1 / float64(factor)
		factor *= 2
	}
	factor = 1
	for i := uint(Places); i < 32; i++ {
		factors[i] = float64(factor)
		factor *= 2
	}
}

// Float64 converts the fixed number to a float64
func (f Fixed) Float64() float64 {
	value := .0
	for _, v := range factors {
		if f&1 == 1 {
			value += v
		}
		f >>= 1
	}
	return value
}

// String converts the fixed point number to a string
func (f Fixed) String() string {
	return fmt.Sprintf("%f", f.Float64())
}

// Mul multiplys to fixed point nuimbers
func (f Fixed) Mul(b Fixed) Fixed {
	return Fixed((int64(f)*int64(b) + FixedHalf) >> Places)
}
