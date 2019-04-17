// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
)

const (
	Chunks    = 8
	ChunkSize = 64
	CASize    = Chunks * ChunkSize
)

type CA struct {
	Rule        uint8
	State, Next []uint64
}

func main() {
	rand.Seed(1)
	ca := CA{
		Rule:  110,
		State: make([]uint64, Chunks),
		Next:  make([]uint64, Chunks),
	}
	for j := range ca.State {
		ca.State[j] = rand.Uint64()
	}
	iterations := 12000
	gray, count := image.NewGray(image.Rect(0, 0, CASize, iterations)), 0
	for i := 0; i < iterations; i++ {
		for _, s := range ca.State {
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
		ca.Step()
		fmt.Printf("iteration: %d\n", i)
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
}

func (ca *CA) Step() {
	rule, state, next := ca.Rule, ca.State, ca.Next
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
			next[out>>6] |= uint64((rule>>index)&0x1) << uint(out&0x3F)
			c--
			out++
		}
	}
	index = ((index << 1) & 0x7) | (state[0] & 0x1)
	next[out>>6] |= uint64((rule>>index)&0x1) << uint(out&0x3F)
	ca.State, ca.Next = next, state
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
