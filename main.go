// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"

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
	Chunks      = 8
	ChunkSize   = 64
	CASize      = Chunks * ChunkSize
	Alpha       = 0.08
	SpikeFactor = 64
	NetworkSize = 10
)

type CA struct {
	Rule                   uint8
	State                  []uint64
	Connections            []int
	On                     uint64
	Low, Complexity, Spike float64
	Note                   uint8
}

func NewCA(rule uint8, size int, rnd *rand.Rand) CA {
	state := make([]uint64, size)
	for j := range state {
		state[j] = rnd.Uint64()
	}
	return CA{
		Rule:        rule,
		State:       state,
		Connections: make([]int, 0, 8),
		Low:         CASize / 2,
	}
}

type Network struct {
	Neurons []CA
	Rnd     *rand.Rand
	Next    []uint64
}

func NewNetwork(seed, size int) Network {
	rnd, neurons := rand.New(rand.NewSource(1)), make([]CA, size)
	for i := range neurons {
		neurons[i] = NewCA(110, Chunks, rnd)
	}
	return Network{
		Neurons: neurons,
		Rnd:     rnd,
		Next:    make([]uint64, Chunks),
	}
}

func (network *Network) Step() {
	neurons, next := network.Neurons, network.Next
	for n := range neurons {
		next = neurons[n].Step(next)
	}
	network.Next = next
}

func (network *Network) Swap(m, n int) {
	a, neurons := network.Rnd.Intn(Chunks), network.Neurons
	neurons[n].State[a], neurons[m].State[a] = neurons[m].State[a], neurons[n].State[a]
}

type Histogram [256]uint64

func (h *Histogram) Entropy() float64 {
	sum := uint64(0)
	for _, v := range h {
		sum += v
	}
	entropy, total := 0.0, float64(sum)
	for _, v := range h {
		if v == 0 {
			continue
		}
		p := float64(v) / total
		entropy += p * math.Log2(p)
	}
	return -entropy
}

type Markov struct {
	Model    [256][256]uint64
	State    uint8
	HasState bool
}

func (m *Markov) Add(symbol uint8) {
	if !m.HasState {
		m.State, m.HasState = symbol, true
		return
	}
	m.Model[m.State][symbol]++
	m.State = symbol
}

func (m *Markov) Entropy() float64 {
	sum, model := uint64(0), &m.Model
	for i := range model {
		for _, v := range model[i] {
			sum += v
		}
	}
	entropy, total := 0.0, float64(sum)
	for i := range model {
		for _, v := range model[i] {
			if v == 0 {
				continue
			}
			p := float64(v) / total
			entropy += p * math.Log2(p)
		}
	}
	return -entropy
}

var options = struct {
	bench *bool
	learn *bool
}{
	bench: flag.Bool("bench", false, "run the test bench"),
	learn: flag.Bool("learn", false, "learn a network"),
}

var Notes = [...]uint8{
	60,
	62,
	64,
	65,
	67,
	69,
	71,
}

func main() {
	flag.Parse()

	if *options.bench {
		bench()
		return
	}

	if *options.learn {
		ga, err := eaopt.NewDefaultGAConfig().NewGA()
		if err != nil {
			panic(err)
		}

		ga.NGenerations = 100
		ga.RNG = rand.New(rand.NewSource(1))
		ga.ParallelEval = true
		ga.PopSize = 100

		ga.Callback = func(ga *eaopt.GA) {
			fmt.Printf("Best fitness at generation %d: %f\n", ga.Generations, ga.HallOfFame[0].Fitness)
			fmt.Println(ga.HallOfFame[0].Genome.(BoolSlice).String())
		}

		err = ga.Minimize(BoolSliceFactory)
		if err != nil {
			panic(err)
		}
		return
	}

	out, err := os.Create("music.midi")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	ticks := smf.MetricTicks(1920)
	wr := mid.NewSMF(out, 1, smfwriter.TimeFormat(ticks))
	wr.TrackSequenceName("music")
	defer wr.EndOfTrack()

	network := NewNetwork(1, 8)
	for i := range network.Neurons {
		network.Neurons[i].AddConnection((i + 7) % 8)
		network.Neurons[i].AddConnection((i + 1) % 8)
	}
	for i, note := range Notes {
		network.Neurons[i].Note = note
	}
	generation := 0
	notes := make([]uint8, 0, 256)
	for generation < 300000 {
		for n := range network.Neurons {
			if r := network.Rnd.Float64() * SpikeFactor; r < network.Neurons[n].Spike {
				m, max := 0, 0.0
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

	maxEntropy := math.Log2(float64(len(Notes)))
	length, maxMarkov := len(notes), 2*maxEntropy
	entropyPoints, markovPoints := make(plotter.XYs, 0, length), make(plotter.XYs, 0, length)
	for i := 0; i < length-63; i++ {
		histogram, markov := Histogram{}, Markov{}
		for j := 0; j < 64; j++ {
			note := notes[i+j]
			histogram[note]++
			markov.Add(note)
		}
		e, m := histogram.Entropy()/maxEntropy, markov.Entropy()/maxMarkov
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
		if r := network.Rnd.Float64() * SpikeFactor; r < network.Neurons[0].Spike {
			network.Swap(0, 1)
			fmt.Printf("fire 0: %d %f\n", i, network.Neurons[0].Spike)
		} else if r < network.Neurons[1].Spike {
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

func (ca *CA) AddConnection(n int) {
	ca.Connections = append(ca.Connections, n)
}

func (ca *CA) Step(next []uint64) []uint64 {
	rule, state, on := ca.Rule, ca.State, uint64(0)
	length := len(state)
	index, out := state[length-1]>>63, 0
	for i, s := range state {
		c := ChunkSize
		if i == 0 {
			index = ((index << 1) & 0x7) | (s & 0x1)
			s >>= 1
			c--
		}
		next[i] = 0
		for c > 0 {
			index = ((index << 1) & 0x7) | (s & 0x1)
			s >>= 1
			bit := uint64((rule >> index) & 0x1)
			on += bit
			next[out>>6] |= bit << uint(out&0x3F)
			c--
			out++
		}
	}
	index = ((index << 1) & 0x7) | (state[0] & 0x1)
	bit := uint64((rule >> index) & 0x1)
	on += bit
	next[out>>6] |= bit << uint(out&0x3F)
	ca.State, ca.On = next, on

	low, complexity := ca.Low, ca.Complexity
	low = low + Alpha*(float64(on)-low)
	complexity = complexity + Alpha*(math.Abs(float64(on)-low)-complexity)
	ca.Low, ca.Complexity, ca.Spike = low, complexity, math.Exp(-complexity)

	return state
}

func (ca *CA) String() string {
	state := ""
	for _, s := range ca.State {
		for i := 0; i < 64; i++ {
			if s&0x1 == 0 {
				state += "0"
			} else {
				state += "1"
			}
			s >>= 1
		}
	}
	return state
}
