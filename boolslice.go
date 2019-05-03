// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"compress/gzip"
	"fmt"
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

type BoolSlice []bool

func (s BoolSlice) At(i int) interface{} {
	return s[i]
}

func (s BoolSlice) Set(i int, v interface{}) {
	s[i] = v.(bool)
}

func (s BoolSlice) Len() int {
	return len(s)
}

func (s BoolSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s BoolSlice) Slice(a, b int) eaopt.Slice {
	return s[a:b]
}

func (s BoolSlice) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return s[:k], s[k:]
}

func (s BoolSlice) Append(t eaopt.Slice) eaopt.Slice {
	return append(s, t.(BoolSlice)...)
}

func (s BoolSlice) Replace(t eaopt.Slice) {
	copy(s, t.(BoolSlice))
}

func (s BoolSlice) Copy() eaopt.Slice {
	t := make(BoolSlice, len(s))
	copy(t, s)
	return t
}

func (s BoolSlice) String() string {
	series, space := "", ""
	for i, value := range s {
		if value {
			series += fmt.Sprintf("%s%d", space, i+1)
			space = " "
		}
	}
	return series
}

type Writer struct {
	Count int
}

func (w *Writer) Write(p []byte) (n int, err error) {
	length := len(p)
	w.Count += length
	return length, nil
}

func (s BoolSlice) Evaluate() (float64, error) {
	network, k := NewNetwork(1, NetworkSize), 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < i; j++ {
			if s[k] {
				network.Neurons[i].AddConnection(j)
				network.Neurons[j].AddConnection(i)
			}
			k++
		}
	}

	network.Neurons[0].Note = 60
	network.Neurons[1].Note = 62
	network.Neurons[2].Note = 64
	network.Neurons[3].Note = 65
	network.Neurons[4].Note = 67
	network.Neurons[5].Note = 69
	network.Neurons[6].Note = 71
	generation := 0
	compressed, markov := Writer{}, Markov{}
	notes, writer := 0, gzip.NewWriter(&compressed)
	for generation < 20000 {
		for n := range network.Neurons {
			if r := network.Rnd.Float64() * SpikeFactor; r < network.Neurons[n].Spike &&
				len(network.Neurons[n].Connections) > 0 {
				m, max := 0, 0.0
				for _, c := range network.Neurons[n].Connections {
					if complexity := network.Neurons[c].Complexity; complexity > max {
						m, max = c, complexity
					}
				}
				network.Swap(n, m)

				if note := network.Neurons[n].Note; note > 0 {
					notes++
					_, err := writer.Write([]byte{note})
					if err != nil {
						panic(err)
					}
					markov.Add(note)
				}
			}
		}
		network.Step()
		generation++
	}

	err := writer.Close()
	if err != nil {
		panic(err)
	}

	//fitness := float64(compressed.Count) / float64(notes)
	//fmt.Println(fitness)
	fitness := markov.Entropy()
	fmt.Println(fitness)
	return fitness, nil
}

func (s BoolSlice) Mutate(rng *rand.Rand) {
	eaopt.MutPermute(s, 1, rng)
}

func (s BoolSlice) Crossover(r eaopt.Genome, rng *rand.Rand) {
	eaopt.CrossGNX(s, r.(BoolSlice), 1, rng)
}

func (s BoolSlice) Clone() eaopt.Genome {
	r := make(BoolSlice, len(s))
	copy(r, s)
	return r
}

func BoolSliceFactory(rnd *rand.Rand) eaopt.Genome {
	s := make(BoolSlice, NetworkSize*(NetworkSize-1)/2)
	k := 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < i; j++ {
			s[k] = rnd.Intn(2) == 0
			k++
		}
	}
	return s
}
