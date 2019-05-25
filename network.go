// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math/rand"
)

// Network is a network of cellular automatons
type Network struct {
	Neurons []CA
	Rnd     *rand.Rand
	Next    []uint64
}

// NewNetwork creates a new network of cellular automatons
func NewNetwork(seed, size int) Network {
	rnd, neurons := rand.New(rand.NewSource(1)), make([]CA, size)
	for i := range neurons {
		neurons[i] = NewCA(110, Chunks, SpikeThreshold, rnd)
	}
	return Network{
		Neurons: neurons,
		Rnd:     rnd,
		Next:    make([]uint64, Chunks),
	}
}

// Step steps all of the cellular automatons in the network
func (network *Network) Step() {
	neurons, next := network.Neurons, network.Next
	for n := range neurons {
		next = neurons[n].Step(next)
	}
	network.Next = next
}

// Swap sends a message between two cellular automatons
func (network *Network) Swap(m, n int) {
	a, b, neurons := network.Rnd.Intn(Chunks), network.Rnd.Intn(Chunks), network.Neurons
	neurons[n].State[a], neurons[m].State[b] = neurons[m].State[b], neurons[n].State[a]
}
