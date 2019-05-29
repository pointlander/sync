// Copyright 2019 The Sync Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"

	"github.com/pointlander/sync/cellular"
	"github.com/pointlander/sync/harmonic"
)

var options = struct {
	bench     *bool
	learn     *bool
	inference *bool
	mode      *string
	net       *string
}{
	bench:     flag.Bool("bench", false, "run the test bench"),
	learn:     flag.Bool("learn", false, "learn a network"),
	inference: flag.Bool("inference", false, "run inference on a network"),
	mode:      flag.String("mode", "harmonic", "harmonic or cellular"),
	net:       flag.String("net", "", "net file to load"),
}

func main() {
	flag.Parse()

	if *options.mode == "cellular" {
		if *options.bench {
			cellular.Bench()
			return
		}

		if *options.learn {
			cellular.Learn()
			return
		}

		if *options.inference {
			cellular.Inference(*options.net)
			return
		}
	} else if *options.mode == "harmonic" {
		if *options.bench {
			harmonic.Bench()
			return
		}

		if *options.learn {
			harmonic.Learn()
			return
		}

		if *options.inference {
			harmonic.Inference(*options.net)
			return
		}
	}
}
