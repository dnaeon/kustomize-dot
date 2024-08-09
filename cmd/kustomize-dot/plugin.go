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
	"bytes"
	"fmt"

	"github.com/dnaeon/kustomize-dot/pkg/parser"
	"github.com/urfave/cli/v2"
	"gopkg.in/dnaeon/go-graph.v1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// pluginConfig contains the plugin configuration
type pluginConfig struct {
	// Spec is the spec of the plugin
	Spec pluginSpec `yaml:"spec"`
}

// pluginSpec contains the config spec for the plugin.
type pluginSpec struct {
	// Layout contains the layout direction
	Layout string `yaml:"layout"`

	// HighlightKinds contains the mapping between Kubernetes resource kind
	// and the color with which to paint it.
	HighlightKinds map[string]string `yaml:"highlightKinds"`

	// HighlightNamespace contains the mapping between Kubernetes namespace
	// and the color with which to paint all resources from the given
	// namespace.
	HighlightNamespaces map[string]string `yaml:"highlightNamespaces"`

	// DropKinds contains the resource kinds which will be dropped from the
	// graph.
	DropKinds []string `yaml:"dropKinds"`

	// DropNamespaces contains the list of namespaces to drop, along with
	// all resources from them.
	DropNamespaces []string `yaml:"dropNamespaces"`

	// KeepKinds contains the list of Kubernetes resources which will be
	// kept. Any other resource will be dropped.
	KeepKinds []string `yaml:"keepKinds"`

	// KeepNamespaces contains the list of namespaces to keep, along with
	// the resources from them. Anything else will be dropped.
	KeepNamespaces []string `yaml:"keepNamespace"`
}

// newPluginCommand returns the command for running kustomize-dot as KRM
// Function plugin
func newPluginCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "plugin",
		Usage:   "run as KRM Function plugin",
		Aliases: []string{"p"},
		Action:  execPluginCommand,
	}

	return cmd
}

// execPluginCommand runs kustomize-dot as a KRM Function plugin.
//
// See [1] for more details about KRM Function plugins.
//
// [1]: https://kubectl.docs.kubernetes.io/guides/extending_kustomize/containerized_krm_functions/
func execPluginCommand(ctx *cli.Context) error {
	var config pluginConfig

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		opts := make([]parser.Option, 0)

		// Layout direction
		opts = append(opts, parser.WithLayoutDirection(parser.LayoutDirection(config.Spec.Layout)))

		// Highlight Resource Kinds
		for kind, color := range config.Spec.HighlightKinds {
			opts = append(opts, parser.WithHighlightKind(kind, color))
		}

		// Highlight Namespaces
		for ns, color := range config.Spec.HighlightNamespaces {
			opts = append(opts, parser.WithHighlightNamespace(ns, color))
		}

		// Drop Resource Kinds
		for _, kind := range config.Spec.DropKinds {
			opts = append(opts, parser.WithDropKind(kind))
		}

		// Drop Namespaces
		for _, ns := range config.Spec.DropNamespaces {
			opts = append(opts, parser.WithDropNamespace(ns))
		}

		// Keep Resource Kinds
		for _, kind := range config.Spec.KeepKinds {
			opts = append(opts, parser.WithKeepKind(kind))
		}

		// Keep Namespaces
		for _, ns := range config.Spec.KeepNamespaces {
			opts = append(opts, parser.WithKeepNamespace(ns))
		}

		// Parse resources and generate the graph
		resources, err := parser.ResourcesFromRNodes(items)
		if err != nil {
			return nil, fmt.Errorf("cannot parse resources: %w", err)
		}

		p := parser.New(opts...)
		g, err := p.Parse(resources)
		if err != nil {
			return nil, fmt.Errorf("cannot generate graph: %w", err)
		}

		var buf bytes.Buffer
		if err := graph.WriteDot(g, &buf); err != nil {
			return nil, err
		}

		// Return the transformed resources as a ConfigMap
		out, err := parser.NewResourceFactory().FromMapWithName(
			"kustomize-dot",
			map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]string{
					"name":      "kustomize-dot",
					"namespace": "default",
				},
				"data": map[string]string{
					"dot": buf.String(),
				},
			},
		)
		if err != nil {
			return nil, err
		}

		return []*yaml.RNode{&out.RNode}, nil
	}

	processor := framework.SimpleProcessor{Config: &config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(processor, command.StandaloneDisabled, false)

	return cmd.Execute()
}
