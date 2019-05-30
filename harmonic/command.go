// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harmonic

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	NetworkSize = 7
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
	MaxEntropy = math.Log2(float64(len(Notes)))
	MaxMarkov  = 2 * MaxEntropy
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

	plots := make([]plotter.XYs, len(network))
	for i := range plots {
		plots[i] = make(plotter.XYs, 0, 1024)
	}
	for i := 0; i < 10000; i++ {
		for j := range network {
			if network[j].Step() {
				fmt.Printf(" %d", network[j].Note)
				plots[j] = append(plots[j], plotter.XY{X: float64(i), Y: network[j].States[0].Float64()})
			}
		}
	}
	fmt.Printf("\n")
	for i := range plots {
		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "state"
		p.X.Label.Text = "time"
		p.Y.Label.Text = "state"

		scatter, err := plotter.NewLine(plots[i])
		if err != nil {
			panic(err)
		}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("harmonic_oscillator_node_%d.png", i))
		if err != nil {
			panic(err)
		}
	}
}
