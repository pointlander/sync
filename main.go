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
)

type CA struct {
	Rule                   uint8
	State                  []uint64
	Connections            []int
	On                     uint64
	Low, Complexity, Spike float64
}

func NewCA(rule uint8, size int) CA {
	state := make([]uint64, size)
	for j := range state {
		state[j] = rand.Uint64()
	}
	return CA{
		Rule:        rule,
		State:       state,
		Connections: make([]int, 0, 8),
		Low:         CASize / 2,
	}
}

var options = struct {
	bench *bool
}{
	bench: flag.Bool("bench", false, "run the test bench"),
}

func main() {
	flag.Parse()

	if *options.bench {
		bench()
		return
	}

	rand.Seed(1)
	nodes, next := make([]CA, 8), make([]uint64, Chunks)
	for i := range nodes {
		nodes[i] = NewCA(110, Chunks)
		nodes[i].AddConnection((i + 7) % 8)
		nodes[i].AddConnection((i + 1) % 8)
	}
	generation := 0
	for {
		for n := range nodes {
			if rnd := rand.Float64() * SpikeFactor; rnd < nodes[n].Spike {
				m, max := 0, 0.0
				for _, c := range nodes[n].Connections {
					if complexity := nodes[c].Complexity; complexity > max {
						m, max = c, complexity
					}
				}
				a := rand.Intn(Chunks)
				nodes[n].State[a], nodes[m].State[a] = nodes[m].State[a], nodes[n].State[a]
				fmt.Printf("fire %d: %d %f\n", n, generation, nodes[n].Spike)
			}
		}
		for n := range nodes {
			next = nodes[n].Step(next)
		}
		generation++
	}
}

func bench() {
	rand.Seed(1)
	iterations, nodes := 12000, make([]CA, 2)
	for i := range nodes {
		nodes[i] = NewCA(110, Chunks)
	}
	points := make(plotter.XYs, 0, iterations)
	gray, count, next := image.NewGray(image.Rect(0, 0, 2*CASize+3, iterations)), 0, make([]uint64, Chunks)
	for i := 0; i < iterations; i++ {
		for n := range nodes {
			for _, s := range nodes[n].State {
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
		if rnd := rand.Float64() * SpikeFactor; rnd < nodes[0].Spike {
			a := rand.Intn(Chunks)
			nodes[0].State[a], nodes[1].State[a] = nodes[1].State[a], nodes[0].State[a]
			fmt.Printf("fire 0: %d %f\n", i, nodes[0].Spike)
		} else if rnd < nodes[1].Spike {
			a := rand.Intn(Chunks)
			nodes[0].State[a], nodes[1].State[a] = nodes[1].State[a], nodes[0].State[a]
			fmt.Printf("fire 1: %d %f\n", i, nodes[1].Spike)
		}
		for n := range nodes {
			next = nodes[n].Step(next)
		}
		points = append(points, plotter.XY{X: float64(i), Y: nodes[0].Spike})
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
