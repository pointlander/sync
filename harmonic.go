// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/gob"
	"math/rand"
	"os"

	"github.com/pointlander/sync/fixed"

	"github.com/MaxHalford/eaopt"
)

// Message is a message sent from one harmonic node to another harmonic node
type Message struct {
	Delay uint8
	Value fixed.Fixed
}

// Channel is an delayed output channel to another harmonic node
type Channel struct {
	Delay  uint8
	Buffer [8]Message
	Out    chan<- fixed.Fixed
}

// Harmonic is a harmonic node
type Harmonic struct {
	Note    uint8
	States  [2]fixed.Fixed
	Weights [4]fixed.Fixed
	Outbox  []Channel
	Inbox   []<-chan fixed.Fixed
}

// HarmonicGenome is a genome representing the parameters of a harmonic network
type HarmonicGenome struct {
	Connections Uint8Slice
	States      FixedSlice
	Weights     FixedSlice
}

// HarmonicNetwork is a network of harmonic nodes
type HarmonicNetwork []Harmonic

// Send sends a delayed message to another harmonic node
func (c *Channel) Send(value fixed.Fixed) {
	if c.Delay == 0 {
		select {
		case c.Out <- value:
		default:
		}
		return
	}
	for i, message := range c.Buffer {
		if message.Delay != 0 {
			continue
		}
		c.Buffer[i] = Message{
			Delay: c.Delay,
			Value: value,
		}
		return
	}
}

// Step steps the state of the channel which can send messages
func (c *Channel) Step() {
	for i, message := range c.Buffer {
		if message.Delay == 0 {
			continue
		}
		message.Delay--
		if message.Delay == 0 {
			select {
			case c.Out <- message.Value:
			default:
			}
		}
		c.Buffer[i] = message
	}
}

// Step steps the state of the harmonic node
func (h *Harmonic) Step() bool {
	outbox := h.Outbox
	for i := range outbox {
		outbox[i].Step()
	}

	sum, count := fixed.Fixed(0), 0
	for _, input := range h.Inbox {
		select {
		case value := <-input:
			sum += value
			count++
		default:
		}
	}

	states, weights := h.States, h.Weights
	states[1], states[0] = states[0], weights[0].Mul(states[0])+weights[1].Mul(states[1])
	if count > 0 {
		states[0] += weights[2].Mul(sum / fixed.Fixed(count))
	}
	fired := false
	if states[0] > weights[3] {
		fired = true
		for i := range outbox {
			outbox[i].Send(states[0])
		}
	}
	h.States = states
	return fired
}

// NewHarmonicNetwork create a harmonic network for a harmonic genome
func (g *HarmonicGenome) NewHarmonicNetwork() HarmonicNetwork {
	network, c, s, w := make(HarmonicNetwork, NetworkSize), 0, 0, 0
	for i := range network {
		for j := range network {
			if delay := g.Connections[c]; i != j && delay < 255 {
				connection := make(chan fixed.Fixed, 8)
				network[i].Outbox = append(network[i].Outbox, Channel{
					Delay: delay,
					Out:   connection,
				})
				network[j].Inbox = append(network[j].Inbox, connection)
			}
			c++
		}
		for j := range network[i].States {
			network[i].States[j] = g.States[s]
			s++
		}
		for j := range network[i].Weights {
			network[i].Weights[j] = g.Weights[w]
			w++
		}
	}
	for i, note := range Notes {
		network[i].Note = note
	}
	return network
}

// Step steps the state of the harmonic network
func (h HarmonicNetwork) Step() (notes []uint8) {
	for i := range h {
		if h[i].Step() {
			notes = append(notes, h[i].Note)
		}
	}
	return notes
}

// Evaluate computes the fitness of the harmonic genome
func (g *HarmonicGenome) Evaluate() (float64, error) {
	network, markov := g.NewHarmonicNetwork(), Markov{}
	for i := 0; i < 10000; i++ {
		notes := network.Step()
		for _, note := range notes {
			markov.Add(note)
		}
	}
	fitness := markov.Entropy()/MaxMarkov - .8
	return fitness * fitness, nil
}

// Mutate mutates the harmonic genome
func (g *HarmonicGenome) Mutate(rng *rand.Rand) {
	eaopt.MutPermute(g.Connections, 1, rng)
	eaopt.MutPermute(g.States, 1, rng)
	eaopt.MutPermute(g.Weights, 1, rng)
}

// Crossover mates two harmonic genomes
func (g *HarmonicGenome) Crossover(r eaopt.Genome, rng *rand.Rand) {
	eaopt.CrossGNX(g.Connections, r.(*HarmonicGenome).Connections, 1, rng)
	eaopt.CrossGNX(g.States, r.(*HarmonicGenome).States, 1, rng)
	eaopt.CrossGNX(g.Weights, r.(*HarmonicGenome).Weights, 1, rng)
}

// Clone produces a copy of a harmonic genome
func (g *HarmonicGenome) Clone() eaopt.Genome {
	connections := make(Uint8Slice, len(g.Connections))
	states := make(FixedSlice, len(g.States))
	weights := make(FixedSlice, len(g.Weights))
	copy(connections, g.Connections)
	copy(states, g.States)
	copy(weights, g.Weights)
	return &HarmonicGenome{
		Connections: connections,
		States:      states,
		Weights:     weights,
	}
}

func (g *HarmonicGenome) Write(name string) {
	out, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	encoder := gob.NewEncoder(out)
	err = encoder.Encode(g)
	if err != nil {
		panic(err)
	}
}

func ReadHarmonicGenome(name string) *HarmonicGenome {
	genome := HarmonicGenome{}
	in, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer in.Close()
	decoder := gob.NewDecoder(in)
	err = decoder.Decode(&genome)
	if err != nil {
		panic(err)
	}
	return &genome
}

// HarmonicGenomeFactory create a new harmonic genome
func HarmonicGenomeFactory(rnd *rand.Rand) eaopt.Genome {
	connections := make(Uint8Slice, NetworkSize*NetworkSize)
	k := 0
	for i := 0; i < NetworkSize; i++ {
		for j := 0; j < NetworkSize; j++ {
			if rnd.Intn(2) == 0 {
				connections[k] = 255
			} else {
				connections[k] = uint8(rnd.Intn(255))
			}
			k++
		}
	}
	states := make(FixedSlice, 2*NetworkSize)
	for i := range states {
		states[i] = fixed.Fixed(rnd.Intn(8 << fixed.Places))
		if rnd.Intn(2) == 0 {
			states[i] = -states[i]
		}
	}
	weights := make(FixedSlice, 4*NetworkSize)
	for i := range weights {
		weights[i] = fixed.Fixed(rnd.Intn(8 << fixed.Places))
		if rnd.Intn(2) == 0 {
			weights[i] = -weights[i]
		}
	}
	return &HarmonicGenome{
		Connections: connections,
		States:      states,
		Weights:     weights,
	}
}
