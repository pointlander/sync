// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"math"
)

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
