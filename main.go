// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	Chunks         = 8
	ChunkSize      = 64
	CASize         = Chunks * ChunkSize
	Alpha          = 0.08
	SpikeFactor    = 2
	SpikeThreshold = .66
	NetworkSize    = 7
)

var options = struct {
	bench     *bool
	learn     *bool
	inference *bool
	net       *string
}{
	bench:     flag.Bool("bench", false, "run the test bench"),
	learn:     flag.Bool("learn", false, "learn a network"),
	inference: flag.Bool("inference", false, "run inference on a network"),
	net:       flag.String("net", "", "net file to load"),
}

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

func bench() {
	network := NewNetwork(1, 2)
	iterations := 12000
	points := make(plotter.XYs, 0, iterations)
	gray, count := image.NewGray(image.Rect(0, 0, 2*CASize+3, iterations)), 0
	for i := 0; i < iterations; i++ {
		for n := range network.Neurons {
			for _, s := range network.Neurons[n].State {
				for j := 0; j < ChunkSize; j++ {
					if s&0x1 == 0 {
						gray.Pix[count] = 0
					} else {
						gray.Pix[count] = 0xFF
					}
					s >>= 1
					count++
				}
			}
			if n == 0 {
				gray.Pix[count] = 0
				count++
				gray.Pix[count] = 0xFF
				count++
				gray.Pix[count] = 0
				count++
			}
		}
		if network.Neurons[0].Test() {
			network.Swap(0, 1)
			fmt.Printf("fire 0: %d %f\n", i, network.Neurons[0].Spike)
		} else if network.Neurons[1].Test() {
			network.Swap(0, 1)
			fmt.Printf("fire 1: %d %f\n", i, network.Neurons[1].Spike)
		}
		network.Step()
		points = append(points, plotter.XY{X: float64(i), Y: network.Neurons[0].Spike})
	}

	out, err := os.Create("ca.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	err = png.Encode(out, gray)
	if err != nil {
		panic(err)
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "complexity"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "complexity"

	scatter, err := plotter.NewScatter(points)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Radius = vg.Length(1)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}
	p.Add(scatter)

	err = p.Save(8*vg.Inch, 8*vg.Inch, "complexity.png")
	if err != nil {
		panic(err)
	}
}

func learn() {
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		panic(err)
	}

	ga.NGenerations = 25
	ga.RNG = rand.New(rand.NewSource(1))
	ga.ParallelEval = true
	ga.PopSize = 100

	ga.Callback = func(ga *eaopt.GA) {
		fmt.Printf("Best fitness at generation %d: %f\n", ga.Generations, ga.HallOfFame[0].Fitness)
		fmt.Println(ga.HallOfFame[0].Genome.(*Net).Connections.String())
	}

	err = ga.Minimize(NetFactory)
	if err != nil {
		panic(err)
	}

	best := ga.HallOfFame[0].Genome.(*Net)
	out, err := os.Create("best.net")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	encoder := gob.NewEncoder(out)
	err = encoder.Encode(best)
	if err != nil {
		panic(err)
	}

}

func inference() {
	out, err := os.Create("music.midi")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	ticks := smf.MetricTicks(1920)
	wr := mid.NewSMF(out, 1, smfwriter.TimeFormat(ticks))
	wr.TrackSequenceName("music")
	defer wr.EndOfTrack()

	network := NewNetwork(1, NetworkSize)
	if *options.net != "" {
		net := Net{}
		in, err := os.Open(*options.net)
		if err != nil {
			panic(err)
		}
		defer in.Close()
		decoder := gob.NewDecoder(in)
		err = decoder.Decode(&net)
		if err != nil {
			panic(err)
		}

		k := 0
		for i := 0; i < NetworkSize; i++ {
			for j := 0; j < NetworkSize; j++ {
				if i != j && net.Connections[k] {
					network.Neurons[i].AddConnection(j)
				}
				k++
			}
		}
		fmt.Println(net.Thresholds)
		for i, value := range net.Thresholds {
			network.Neurons[i].Threshold = value
		}
	} else {
		for i := range network.Neurons {
			network.Neurons[i].AddConnection((i + (NetworkSize - 1)) % NetworkSize)
			network.Neurons[i].AddConnection((i + 1) % NetworkSize)
		}
	}
	for i, note := range Notes {
		network.Neurons[i].Note = note
	}

	generation := 0
	notes := make([]uint8, 0, 256)
	for generation < 300000 {
		for n := range network.Neurons {
			if network.Neurons[n].Test() {
				m, max := n, 0.0
				for _, c := range network.Neurons[n].Connections {
					if complexity := network.Neurons[c].Complexity; complexity > max {
						m, max = c, complexity
					}
				}
				network.Swap(n, m)
				fmt.Printf("fire %d: %d %f\n", n, generation, network.Neurons[n].Spike)

				if note := network.Neurons[n].Note; note > 0 {
					wr.SetDelta(ticks.Ticks8th())
					wr.NoteOn(note, 50)
					wr.SetDelta(ticks.Ticks8th())
					wr.NoteOff(note)
					notes = append(notes, note)
				}
			}
		}
		network.Step()
		generation++
	}

	length := len(notes)
	entropyPoints, markovPoints := make(plotter.XYs, 0, length), make(plotter.XYs, 0, length)
	for i := 0; i < length-63; i++ {
		histogram, markov := Histogram{}, Markov{}
		for j := 0; j < 64; j++ {
			note := notes[i+j]
			histogram[note]++
			markov.Add(note)
		}
		e, m := histogram.Entropy()/MaxEntropy, markov.Entropy()/MaxMarkov
		entropyPoints = append(entropyPoints, plotter.XY{X: float64(i), Y: e})
		markovPoints = append(markovPoints, plotter.XY{X: float64(i), Y: m})
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "entropy"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "entrpy"

	scatter, err := plotter.NewScatter(entropyPoints)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Radius = vg.Length(1)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}
	p.Add(scatter)

	err = p.Save(8*vg.Inch, 8*vg.Inch, "entropy.png")
	if err != nil {
		panic(err)
	}

	p, err = plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "markov"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "markov"

	scatter, err = plotter.NewScatter(markovPoints)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Radius = vg.Length(1)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}
	p.Add(scatter)

	err = p.Save(8*vg.Inch, 8*vg.Inch, "markov.png")
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if *options.bench {
		bench()
		return
	}

	if *options.learn {
		learn()
		return
	}

	if *options.inference {
		inference()
		return
	}

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
