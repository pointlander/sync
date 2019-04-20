// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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
	Chunks    = 8
	ChunkSize = 64
	CASize    = Chunks * ChunkSize
)

type CA struct {
	Rule    uint8
	State   []uint64
	Entropy float64
}

func NewCA(rule uint8, size int) CA {
	state := make([]uint64, size)
	for j := range state {
		state[j] = rand.Uint64()
	}
	return CA{
		Rule:  rule,
		State: state,
	}
}

func main() {
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
		if rnd := rand.Float64() * 32; rnd < nodes[0].Entropy || rnd < nodes[1].Entropy {
			//a, b := rand.Intn(Chunks), rand.Intn(Chunks)
			//nodes[0].State[a], nodes[1].State[b] = nodes[1].State[b], nodes[0].State[a]
			a := rand.Intn(Chunks)
			nodes[0].State[a], nodes[1].State[a] = nodes[1].State[a], nodes[0].State[a]
			fmt.Printf("fire: %d %f %f\n", i, nodes[0].Entropy, nodes[1].Entropy)
		}
		for n := range nodes {
			next = nodes[n].Step(next)
		}
		points = append(points, plotter.XY{X: float64(i), Y: nodes[0].Entropy})
		//fmt.Printf("iteration: %d %f %f\n", i, nodes[0].Entropy, nodes[1].Entropy)
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

	p.Title.Text = "entropy"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "entropy"

	scatter, err := plotter.NewScatter(points)
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
}

func (ca *CA) Step(next []uint64) []uint64 {
	rule, state, histogram := ca.Rule, ca.State, [8]uint64{}
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
			histogram[index]++
			s >>= 1
			next[out>>6] |= uint64((rule>>index)&0x1) << uint(out&0x3F)
			c--
			out++
		}
	}
	index = ((index << 1) & 0x7) | (state[0] & 0x1)
	histogram[index]++
	next[out>>6] |= uint64((rule>>index)&0x1) << uint(out&0x3F)
	ca.State = next

	entropy := 0.0
	for _, i := range histogram {
		p := float64(i) / 512
		entropy += p * math.Log2(p)
	}
	ca.Entropy = math.Exp(entropy)
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
