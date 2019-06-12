// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harmonic

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
	"github.com/mjibson/go-dsp/fft"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	NetworkSize = 7
	Iterations  = 10000
)

var (
	Notes = [...]uint8{
		60,
		62,
		64,
		65,
		67,
		69,
		71,
	}
	MaxEntropy         = math.Log2(float64(len(Notes)))
	MaxMarkov          = 2 * MaxEntropy
	MaxSpectrumEntropy = math.Log(Iterations)
)

func Bench() {
	points := make(plotter.XYs, 0, 1024)

	Fs := 44100.0
	f0 := 20.0
	x0 := 3.0
	v0 := 0.6
	T := 1 / Fs
	w0 := 2 * math.Pi * f0
	if T >= 2/w0 {
		panic("This is unstable")
	}
	coefficient1 := 2 - (T*T)*(w0*w0)

	x1 := fixed.FixedFloat64(x0)
	x2 := fixed.FixedFloat64(x0 + T*v0)
	c := fixed.FixedFloat64(coefficient1)
	fmt.Printf("%s %s %s\n", x1, x2, c)
	for i := 0; i < 10000; i++ {
		x1, x2 = x2, c.Mul(x2)-x1
		points = append(points, plotter.XY{X: float64(i), Y: x2.Float64()})
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "state"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "state"

	scatter, err := plotter.NewLine(points)
	if err != nil {
		panic(err)
	}
	p.Add(scatter)

	err = p.Save(8*vg.Inch, 8*vg.Inch, "harmonic_oscillator.png")
	if err != nil {
		panic(err)
	}
}

func Learn() {
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		panic(err)
	}

	ga.NGenerations = 200
	ga.RNG = rand.New(rand.NewSource(1))
	ga.ParallelEval = true
	ga.PopSize = 100

	ga.Callback = func(ga *eaopt.GA) {
		fmt.Printf("Best fitness at generation %d: %f\n", ga.Generations, ga.HallOfFame[0].Fitness)
		fmt.Println(ga.HallOfFame[0].Genome.(*HarmonicGenome).Connections.String())
	}
	ga.EarlyStop = func(ga *eaopt.GA) bool {
		return ga.HallOfFame[0].Fitness < 0.00001
	}

	err = ga.Minimize(HarmonicGenomeFactory)
	if err != nil {
		panic(err)
	}

	best := ga.HallOfFame[0].Genome.(*HarmonicGenome)
	best.Write("best_harmonic.net")
}

func Entropy(values []complex128) (entropy float64) {
	length, total := len(values), 0.0
	probability, n := make([]float64, 0, length), float64(length)
	for _, value := range values {
		a, b := real(value), imag(value)
		p := (a*a + b*b) / n
		total += p
		probability = append(probability, p)
	}
	for _, p := range probability {
		p /= total
		entropy += p * math.Log(p)
	}
	return -entropy
}

func Inference(name string) {
	if name == "" {
		panic("net file required")
	}
	genome := ReadHarmonicGenome(name)
	network := genome.NewHarmonicNetwork()
	for i := range network {
		fmt.Println(i)
		fmt.Println(network[i].States)
		fmt.Println(network[i].Weights)
	}

	data := make([][]float64, len(network))
	for i := range data {
		data[i] = make([]float64, 0, Iterations)
	}
	for i := 0; i < Iterations; i++ {
		notes := network.Step(data)
		for _, note := range notes {
			fmt.Printf(" %d", note)
		}
	}
	fmt.Printf("\n")

	points := make(plotter.XYs, Iterations)
	for i, values := range data {
		fmt.Printf("graphing plot %d\n", i)

		for j, value := range values {
			points[j] = plotter.XY{
				X: float64(j),
				Y: value,
			}
		}

		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "state"
		p.X.Label.Text = "time"
		p.Y.Label.Text = "state"

		scatter, err := plotter.NewScatter(points)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("harmonic_oscillator_node_%d.png", i))
		if err != nil {
			panic(err)
		}

		spectrum := fft.FFTReal(values)
		for j, value := range spectrum {
			points[j] = plotter.XY{
				X: float64(j),
				Y: cmplx.Abs(value),
			}
		}
		entropy := Entropy(spectrum)
		fmt.Println("entopy=", entropy/MaxSpectrumEntropy)

		p, err = plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "spectrum"
		p.X.Label.Text = "frequency"
		p.Y.Label.Text = "energy"

		scatter, err = plotter.NewScatter(points)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("harmonic_oscillator_node_%d_spectrum.png", i))
		if err != nil {
			panic(err)
		}
	}
}
