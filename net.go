// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

type Net struct {
	Connections BoolSlice
	Thresholds  Float64Slice
}

func (n *Net) fitness(seed int) float64 {
	network, k := NewNetwork(seed, NetworkSize), 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < NetworkSize; j++ {
			if i != j && n.Connections[k] {
				network.Neurons[i].AddConnection(j)
			}
			k++
		}
	}
	for i, value := range n.Thresholds {
		network.Neurons[i].Threshold = value
	}
	for i, note := range Notes {
		network.Neurons[i].Note = note
	}

	markov := Markov{}
	for generation := 0; generation < 40000; generation++ {
		for n := range network.Neurons {
			if network.Neurons[n].Test() {
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

func (n *Net) Evaluate() (float64, error) {
	fitness := (n.fitness(1)+n.fitness(2))/2 - .8
	//fmt.Println(fitness)
	return fitness * fitness, nil
}

func (n *Net) Mutate(rng *rand.Rand) {
	eaopt.MutPermute(n.Connections, 1, rng)
	eaopt.MutPermute(n.Thresholds, 1, rng)
}

func (n *Net) Crossover(r eaopt.Genome, rng *rand.Rand) {
	eaopt.CrossGNX(n.Connections, r.(*Net).Connections, 1, rng)
	eaopt.CrossGNX(n.Thresholds, r.(*Net).Thresholds, 1, rng)
}

func (n *Net) Clone() eaopt.Genome {
	connections := make(BoolSlice, len(n.Connections))
	thresholds := make(Float64Slice, len(n.Thresholds))
	copy(connections, n.Connections)
	copy(thresholds, n.Thresholds)
	return &Net{
		Connections: connections,
		Thresholds:  thresholds,
	}
}

func NetFactory(rnd *rand.Rand) eaopt.Genome {
	connections := make(BoolSlice, NetworkSize*NetworkSize)
	k := 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < NetworkSize; j++ {
			connections[k] = rnd.Intn(2) == 0
			k++
		}
	}
	thresholds := make(Float64Slice, NetworkSize)
	for i := range thresholds {
		thresholds[i] = rnd.Float64()
	}
	return &Net{
		Connections: connections,
		Thresholds:  thresholds,
	}
}
