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

// Package parser provides utilities for generating dependency graphs by parsing
// Kubernetes resources produced by kustomize build.
package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/dnaeon/go-graph.v1"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// notClonedPrefix is the prefix added by kustomize for the origin annotation,
// which will be stripped when we generate the graph.
const notClonedPrefix = "notCloned/"

// LayoutDirection is a type which represents the direction of the graph layout.
type LayoutDirection string

// String implements the [fmt.Stringer] interface
func (ld LayoutDirection) String() string {
	return string(ld)
}

const (
	// LayoutDirectionTB specifies Top-to-Botton layout
	LayoutDirectionTB LayoutDirection = "TB"

	// LayoutDirectionBT specifies Botton-to-Top layout
	LayoutDirectionBT LayoutDirection = "BT"

	// LayoutDirectionLR specifies Left-to-Right layout
	LayoutDirectionLR LayoutDirection = "LR"

	// LayoutDirectionRL specifies Right-to-Left layout
	LayoutDirectionRL LayoutDirection = "RL"
)

// NewDepProvider creates a new [provider.DepProvider].
func NewDepProvider() *provider.DepProvider {
	return provider.NewDefaultDepProvider()
}

// NewResourceFactory creates a new [resource.Factory]
func NewResourceFactory() *resource.Factory {
	return NewDepProvider().GetResourceFactory()
}

// ResourcesFromBytes returns the list of [resource.Resource] items contained
// within the given data.
func ResourcesFromBytes(data []byte) ([]*resource.Resource, error) {
	return NewResourceFactory().SliceFromBytes(data)
}

// ResourcesFromRNodes returns the list of [resource.Resource] items represented
// by the given sequence of [yaml.RNode] items.
func ResourcesFromRNodes(items []*yaml.RNode) ([]*resource.Resource, error) {
	return NewResourceFactory().ResourcesFromRNodes(items)
}

// ResourcesFromPath returns the list of [resource.Resource] items by parsing
// the Kubernetes resources from the given path.
func ResourcesFromPath(path string) ([]*resource.Resource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ResourcesFromBytes(data)
}

// ResourcesFromReader returns the list of [resource.Resource] items by parsing
// the Kubernetes resource from the given [io.Reader].
func ResourcesFromReader(r io.Reader) ([]*resource.Resource, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return ResourcesFromBytes(data)
}

// Parser knows how to parse a sequence of [resource.Resource] items and build a
// dependency graph out of it.
//
// The vertices in the graph represent the [resource.Resource] instances, which
// are connected via edges to their origins.
type Parser struct {
	// highlightKindMap contains mappings between Kubernetes resource kinds
	// and the color with which to paint resources with the respective kind.
	highlightKindMap map[string]string

	// highlightNamespaceMap contains the mapping between Kubernetes
	// namespaces and the color with which to paint all resources from the
	// respective namespace.
	highlightNamespaceMap map[string]string

	// layoutDirection specifies the direction of graph layout.
	layoutDirection LayoutDirection

	// dropResourceKinds contains the list of resource kinds, which will be
	// dropped from the resulting graph.
	dropResourceKinds []string

	// dropNamespaces contains the list of namespaces, from which
	// any resource in the specified namespaces will be dropped from the
	// resulting graph.
	dropNamespaces []string

	// keepResourceKinds contains the list of resource kinds, which will be
	// kept. Any other resource kind will be dropped from the resulting
	// graph.
	keepResourceKinds []string

	// keepNamespaces contains the list of namespaces, from which resources
	// will be kept. Any resource, which is not in the specified namespaces
	// will be dropped.
	keepNamespaces []string
}

// New creates a new [Parser] and configures it using the specified options.
func New(opts ...Option) *Parser {
	p := &Parser{
		highlightKindMap:      make(map[string]string),
		highlightNamespaceMap: make(map[string]string),
		layoutDirection:       LayoutDirectionLR,
		dropResourceKinds:     make([]string, 0),
		dropNamespaces:        make([]string, 0),
		keepResourceKinds:     make([]string, 0),
		keepNamespaces:        make([]string, 0),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option is a function which configures the [Parser].
type Option func(p *Parser)

// WithHighlightKind is an [Option] which configures the [Parser] to paint resources
// with the given Kubernetes Resource Kind with the specified color.
func WithHighlightKind(kind string, color string) Option {
	opt := func(p *Parser) {
		p.highlightKindMap[strings.ToLower(kind)] = color
	}

	return opt
}

// WithHighlightNamespace is an [Option] which configures the [Parser] to paint
// all resources from the given namespace with the specified color.
func WithHighlightNamespace(namespace string, color string) Option {
	opt := func(p *Parser) {
		p.highlightNamespaceMap[strings.ToLower(namespace)] = color
	}

	return opt
}

// WithLayoutDirection is an [Option] which configures the [Parser] to generate
// the graph with the specified direction.
func WithLayoutDirection(layout LayoutDirection) Option {
	opt := func(p *Parser) {
		p.layoutDirection = layout
	}

	return opt
}

// WithDropKind is an [Option], which configures the [Parser] to drop the
// specified Kubernetes resource kind from the resulting graph.
func WithDropKind(kind string) Option {
	opt := func(p *Parser) {
		p.dropResourceKinds = append(p.dropResourceKinds, strings.ToLower(kind))
	}

	return opt
}

// WithKeepKind is an [Option], which configures the [Parser] to keep only
// resources of the given kind. Any other resource kind will be dropped from the
// resulting graph.
func WithKeepKind(kind string) Option {
	opt := func(p *Parser) {
		p.keepResourceKinds = append(p.keepResourceKinds, strings.ToLower(kind))
	}

	return opt
}

// WithDropNamespace is an [Option], which configures the [Parser] to drop all
// resources from the specified namespace.
func WithDropNamespace(namespace string) Option {
	opt := func(p *Parser) {
		p.dropNamespaces = append(p.dropNamespaces, strings.ToLower(namespace))
	}

	return opt
}

// WithKeepNamespace is an [Option], which configures the [Parser] to keep only
// resources from the specified namespace. Any other resource will be dropped
// from the resulting graph.
func WithKeepNamespace(namespace string) Option {
	opt := func(p *Parser) {
		p.keepNamespaces = append(p.keepNamespaces, strings.ToLower(namespace))
	}

	return opt
}

// Parse parses the given sequence of [resource.Resource] items in order to
// generate a directed [graph.Graph].
func (p *Parser) Parse(resources []*resource.Resource) (graph.Graph[string], error) {
	g := graph.New[string](graph.KindDirected)

	for _, r := range resources {
		if p.shouldDropResource(r) {
			continue
		}

		// Add u to the graph, and paint the vertex
		uName := p.vertexNameFromResource(r)
		u := g.AddVertex(uName)
		p.applyHighlights(u, r)

		// Add v to the graph, which represents the resource origin
		origin, err := r.GetOrigin()
		if err != nil {
			return nil, err
		}

		// No origin metadata found, skip it
		if origin == nil {
			continue
		}

		vName := p.vertexNameFromOrigin(origin)
		g.AddVertex(vName)

		e := g.AddEdge(uName, vName)
		label := p.edgeLabelFromOrigin(origin)
		e.DotAttributes["label"] = label
	}

	// Set direction of graph layout and other graph-specific attributes
	graphAttrs := g.GetDotAttributes()
	graphAttrs["rankdir"] = p.layoutDirection.String()

	return g, nil
}

// shouldDropResource is a predicate, which returns true, if the resource is to
// be dropped from the graph, and returns false otherwise.
func (p *Parser) shouldDropResource(r *resource.Resource) bool {
	kind := strings.ToLower(r.GetKind())
	namespace := strings.ToLower(r.GetNamespace())

	// Drop resource, if it is part of any drop-namespaces
	for _, dn := range p.dropNamespaces {
		if namespace == dn {
			return true
		}
	}

	// Drop resource, if it is part of any drop-resource-kinds
	for _, drk := range p.dropResourceKinds {
		if kind == drk {
			return true
		}
	}

	// Drop resources, if they are outside of the configured keep-namespaces
	keepNamespaceIsSet := false
	keepKindIsSet := false
	foundKeepNamespace := false
	foundKeepKind := false
	if len(p.keepNamespaces) > 0 {
		keepNamespaceIsSet = true
		for _, kn := range p.keepNamespaces {
			if namespace == kn {
				foundKeepNamespace = true
				break
			}
		}
	}

	// Drop resources, if they are not part of the configured
	// keep-resource-kinds.
	if len(p.keepResourceKinds) > 0 {
		keepKindIsSet = true
		for _, krk := range p.keepResourceKinds {
			if kind == krk {
				foundKeepKind = true
				break
			}
		}
	}

	switch {
	case keepNamespaceIsSet && !foundKeepNamespace:
		// Resource is not part of the keep-namespaces, so drop it.
		return true
	case keepNamespaceIsSet && keepKindIsSet && foundKeepNamespace && !foundKeepKind:
		// Resource is part of the keep-namespaces, but not part of the
		// keep-resource-kinds, so drop it.
		return true
	case keepKindIsSet && !foundKeepKind:
		// Resource is not part of the keep-resource-kinds, so drop it
		return true
	default:
		// Don't drop the resource
		return false
	}
}

// applyHighlights applies the highlight styles to the [graph.Vertex] u for
// [resource.Resource] r.
func (p *Parser) applyHighlights(u *graph.Vertex[string], r *resource.Resource) {
	// First we paint resources by namespace
	namespace := strings.ToLower(r.GetNamespace())
	kind := strings.ToLower(r.GetKind())

	namespaceColor, ok := p.highlightNamespaceMap[namespace]
	if ok {
		u.DotAttributes["color"] = namespaceColor
		u.DotAttributes["fillcolor"] = namespaceColor
	}

	// Then we paint resources by kind
	kindColor, ok := p.highlightKindMap[kind]
	if ok {
		u.DotAttributes["color"] = kindColor
		u.DotAttributes["fillcolor"] = kindColor
	}
}

// vertexNameFromResource returns a string representing the vertex name for the
// given [resource.Resource].
func (p *Parser) vertexNameFromResource(r *resource.Resource) string {
	namespace := r.GetNamespace()
	name := r.GetName()
	kind := strings.ToLower(r.GetKind())

	// Cluster-scoped resource
	if namespace == "" {
		return fmt.Sprintf("%s/%s", kind, name)
	}

	// Namespace-scoped resource
	return fmt.Sprintf("%s/%s/%s", namespace, kind, name)
}

// vertexNameFromOrigin returns a string representing the vertex name for the
// given [resource.Origin].
func (p *Parser) vertexNameFromOrigin(origin *resource.Origin) string {
	path := origin.Path
	path = strings.TrimPrefix(path, notClonedPrefix)
	switch {
	case origin.ConfiguredIn != "":
		// Generator or transformer created resource
		return origin.ConfiguredIn
	case origin.Repo != "":
		// Remote resource
		return path
	default:
		// Local resource
		return path
	}
}

// edgeLabelFromOrigin returns a string to be used as an edge label.
func (p *Parser) edgeLabelFromOrigin(origin *resource.Origin) string {
	switch {
	case origin.ConfiguredIn != "":
		// Generator or transformer created resource
		return fmt.Sprintf("%s/%s", origin.ConfiguredBy.APIVersion, origin.ConfiguredBy.Kind)
	case origin.Repo != "":
		// Remote resource
		if origin.Ref != "" {
			return fmt.Sprintf("%s (ref %s)", origin.Repo, origin.Ref)
		}
		return origin.Repo
	default:
		// Local resource
		return ""
	}
}
