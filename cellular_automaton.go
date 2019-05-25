// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"math/rand"
)

// CA is a cellular automaton
type CA struct {
	Rule                   uint8
	State                  []uint64
	Connections            []int
	On                     uint64
	Low, Complexity, Spike float64
	Threshold              float64
	Note                   uint8
}

// NewCA creates a new cellular automaton
func NewCA(rule uint8, size int, threshold float64, rnd *rand.Rand) CA {
	state := make([]uint64, size)
	for j := range state {
		state[j] = rnd.Uint64()
	}
	return CA{
		Rule:        rule,
		State:       state,
		Connections: make([]int, 0, 8),
		Low:         CASize / 2,
		Threshold:   threshold,
	}
}

// AddConnection adds a connection to another cellular automaton
func (ca *CA) AddConnection(n int) {
	ca.Connections = append(ca.Connections, n)
}

// Test checks if the cellular automaton is firing
func (ca *CA) Test() bool {
	return ca.Spike > ca.Threshold/SpikeFactor
}

// Step generates the next step of the cellular automaton
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

// String converts the cellular automaton to a string
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
