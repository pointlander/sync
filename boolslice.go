// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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

func (s BoolSlice) fitness(seed int) float64 {
	network, k := NewNetwork(seed, NetworkSize), 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < NetworkSize; j++ {
			if i != j && s[k] {
				network.Neurons[i].AddConnection(j)
			}
			k++
		}
	}
	for i, note := range Notes {
		network.Neurons[i].Note = note
	}

	markov := Markov{}
	for generation := 0; generation < 40000; generation++ {
		for n := range network.Neurons {
			if network.Neurons[n].Spike > SpikeThreshold {
				m, max := n, 0.0
				for _, c := range network.Neurons[n].Connections {
					if complexity := network.Neurons[c].Complexity; complexity > max {
						m, max = c, complexity
					}
				}
				network.Swap(n, m)

				if note := network.Neurons[n].Note; note > 0 {
					markov.Add(note)
				}
			}
		}
		network.Step()
	}

	return markov.Entropy() / MaxMarkov
}

func (s BoolSlice) Evaluate() (float64, error) {
	fitness := (s.fitness(1)+s.fitness(2))/2 - .8
	//fmt.Println(fitness)
	return fitness * fitness, nil
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
	s := make(BoolSlice, NetworkSize*NetworkSize)
	k := 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < NetworkSize; j++ {
			s[k] = rnd.Intn(2) == 0
			k++
		}
	}
	return s
}
