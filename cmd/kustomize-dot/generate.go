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
	"io"
	"os"

	"github.com/dnaeon/kustomize-dot/pkg/parser"
	"github.com/urfave/cli/v2"
	"gopkg.in/dnaeon/go-graph.v1"
	"sigs.k8s.io/kustomize/api/resource"
)

// newGenerateCommand returns the command for generating dot representation of
// the Kubernetes resources.
func newGenerateCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "generate",
		Usage:   "generate dot representation",
		Aliases: []string{"gen", "g"},
		Action:  execGenerateCommand,
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
				Aliases: []string{"kind-color", "hk"},
				EnvVars: []string{"HIGHLIGHT_KIND", "KIND_COLOR"},
			},
			&cli.StringSliceFlag{
				Name:    "highlight-namespace",
				Usage:   "highlight resources from a given namespace with specified color",
				Aliases: []string{"namespace-color", "hn"},
				EnvVars: []string{"HIGHLIGHT_NAMESPACE", "NAMESPACE_COLOR"},
			},
			&cli.StringSliceFlag{
				Name:    "drop-kind",
				Usage:   "drop resources of the given kind",
				Aliases: []string{"dk"},
				EnvVars: []string{"DROP_KIND"},
			},
			&cli.StringSliceFlag{
				Name:    "drop-namespace",
				Usage:   "drop all resources from the given namespace",
				Aliases: []string{"dn"},
				EnvVars: []string{"DROP_NAMESPACE"},
			},
			&cli.StringSliceFlag{
				Name:    "keep-kind",
				Usage:   "keep resources of the given kind only",
				Aliases: []string{"kk"},
				EnvVars: []string{"KEEP_KIND"},
			},
			&cli.StringSliceFlag{
				Name:    "keep-namespace",
				Usage:   "keep resources from the given namespace only",
				Aliases: []string{"kn"},
				EnvVars: []string{"KEEP_NAMESPACE"},
			},
		},
	}

	return cmd
}

// execGenerateCommand runs the command for generating dot representation of the
// Kubernetes resources.
func execGenerateCommand(ctx *cli.Context) error {
	layout, err := getLayoutDirection(ctx)
	if err != nil {
		return err
	}

	// graph layout direction
	opts := make([]parser.Option, 0)
	opts = append(opts, parser.WithLayoutDirection(layout))

	// highlight-kind options
	hkValues := ctx.StringSlice("highlight-kind")
	hkPairs, err := parseKV(hkValues...)
	if err != nil {
		return err
	}
	for _, pair := range hkPairs {
		opts = append(opts, parser.WithHighlightKind(pair.key, pair.val))
	}

	// highlight-namespace options
	hnValues := ctx.StringSlice("highlight-namespace")
	hnPairs, err := parseKV(hnValues...)
	if err != nil {
		return err
	}
	for _, pair := range hnPairs {
		opts = append(opts, parser.WithHighlightNamespace(pair.key, pair.val))
	}

	// drop-kind options
	dkValues := ctx.StringSlice("drop-kind")
	for _, dk := range dkValues {
		opts = append(opts, parser.WithDropKind(dk))
	}

	// drop-namespace options
	dnValues := ctx.StringSlice("drop-namespace")
	for _, dn := range dnValues {
		opts = append(opts, parser.WithDropNamespace(dn))
	}

	// keep-kind options
	kkValues := ctx.StringSlice("keep-kind")
	for _, kk := range kkValues {
		opts = append(opts, parser.WithKeepKind(kk))
	}

	// keep-namespace options
	knValues := ctx.StringSlice("keep-namespace")
	for _, kn := range knValues {
		opts = append(opts, parser.WithKeepNamespace(kn))
	}

	// Read the resources and generate the graph
	var resources []*resource.Resource

	file := ctx.Path("file")
	if file == "-" {
		// Special case for resources passed on stdin
		data, err := io.ReadAll(os.Stdin)
		resources, err = parser.ResourcesFromBytes(data)
		if err != nil {
			return err
		}
	} else {
		resources, err = parser.ResourcesFromPath(file)
		if err != nil {
			return err
		}
	}

	p := parser.New(opts...)
	g, err := p.Parse(resources)
	if err != nil {
		return err
	}

	return graph.WriteDot(g, os.Stdout)
}
