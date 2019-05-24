package fixed

import (
	"testing"
)

func TestFixed_Mul(t *testing.T) {
	a := Fixed(1 << 5)
	b := a.Mul(a)
	if b.Float64() != .25 {
		t.Fatalf("%s != .25", b)
	}
}

func BenchmarkFixedMul(t *testing.B) {
	a, b, c := make([]Fixed, t.N), make([]Fixed, t.N), make([]Fixed, t.N)
	x, y := Fixed(1<<6), Fixed(1<<5)
	for i := 0; i < t.N; i++ {
		a[i], b[i] = x, y
		x++
		y++
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		c[i] = a[i].Mul(b[i])
	}
}

func BenchmarkMulFloat32(t *testing.B) {
	a, b, c := make([]float32, t.N), make([]float32, t.N), make([]float32, t.N)
	x, y := float32(1.0), float32(.5)
	for i := 0; i < t.N; i++ {
		a[i], b[i] = x, y
		x += 1 / 64
		y += 1 / 64
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		c[i] = a[i] * b[i]
	}
}

func BenchmarkMulFloat64(t *testing.B) {
	a, b, c := make([]float64, t.N), make([]float64, t.N), make([]float64, t.N)
	x, y := 1.0, .5
	for i := 0; i < t.N; i++ {
		a[i], b[i] = x, y
		x += 1 / 64
		y += 1 / 64
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		c[i] = a[i] * b[i]
	}
}
