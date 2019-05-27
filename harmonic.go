// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "github.com/pointlander/sync/fixed"

type Message struct {
	Delay uint8
	Value fixed.Fixed
}

type Channel struct {
	Delay  uint8
	Buffer [8]Message
	Out    chan<- fixed.Fixed
}

type Harmonic struct {
	States  [2]fixed.Fixed
	Weights [4]fixed.Fixed
	Outbox  []Channel
	Inbox   []<-chan fixed.Fixed
}

func (c *Channel) Send(value fixed.Fixed) {
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

func (h *Harmonic) Step() {
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
	states[1], states[0] = states[0], weights[0].Mul(states[0])+weights[1].Mul(states[1])+weights[3]
	if count > 0 {
		states[0] += weights[2].Mul(sum / fixed.Fixed(count))
	}
	if states[0] > 0 {
		for i := range outbox {
			outbox[i].Send(states[0])
		}
	}
	h.States = states
}
