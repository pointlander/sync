package fixed

import "testing"

func TestFixed_Mul(t *testing.T) {
	a := Fixed(1 << 5)
	b := a.Mul(a)
	if b.Float64() != .25 {
		t.Fatalf("%s != .25", b)
	}
}

var z Fixed

func BenchmarkFixedMul(t *testing.B) {
	x, y := Fixed(1<<6), Fixed(1<<5)
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
