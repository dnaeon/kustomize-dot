// Copyright (c) 2024 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//   1. Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//   2. Redistributions in binary form must reproduce the above copyright
//      notice, this list of conditions and the following disclaimer in the
//      documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "kustomize-dot",
		Version:              "0.1.0",
		EnableBashCompletion: true,
		Usage:                "tool for generating graphs from kustomize resources",
		Authors: []*cli.Author{
			{
				Name:  "Marin Atanasov Nikolov",
				Email: "dnaeon@gmail.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "layout",
				Usage:   "direction of graph layout",
				Value:   "LR",
				Aliases: []string{"l"},
			},
			&cli.PathFlag{
				Name:     "file",
				Usage:    "file containing the Kubernetes resources",
				Required: true,
				Aliases:  []string{"f"},
			},
			&cli.StringSliceFlag{
				Name:    "highlight-kind",
				Usage:   "highlight resources of a given kind with specified color",
				Aliases: []string{"kind-color"},
			},
			&cli.StringSliceFlag{
				Name:    "highlight-namespace",
				Usage:   "highlight resources from a given namespace with specified color",
				Aliases: []string{"namespace-color"},
			},
			&cli.StringSliceFlag{
				Name:  "drop-kind",
				Usage: "drop resources of a given kind",
			},
			&cli.StringSliceFlag{
				Name:  "drop-namespace",
				Usage: "drop all resources from a given namespace",
			},
			&cli.StringSliceFlag{
				Name:  "keep-kind",
				Usage: "keep resources of a given kind only",
			},
			&cli.StringSliceFlag{
				Name:  "keep-namespace",
				Usage: "keep resources from a given namespace only",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
