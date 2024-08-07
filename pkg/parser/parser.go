// Package parser provides utilities for generating dependency graphs by parsing
// Kubernetes resources produced by `kustomize build'.

package parser

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/dnaeon/go-graph.v1"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

// Parser knows how to parse a sequence of [resource.Resource] items and build a
// dependency graph out of it.
//
// The vertices in the graph represent the [resource.Resource] instances, which
// are connected via edges to their origins.
type Parser struct {
	// highlightKindMap contains mappings between Kubernetes resource kinds
	// and the color with which to paint resources of such kind.
	highlightKindMap map[string]string
}

// New creates a new [Parser] and configures it using the specified options.
func New(opts ...Option) *Parser {
	p := &Parser{
		highlightKindMap: make(map[string]string),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option is a function which configures the [Parser].
type Option func(p *Parser)

// WithHighlight is an [Option] which configures the [Parser] to highlight
// resources with the specified Kubernetes Resource kind and color.
func WithHighlightKind(kind string, color string) Option {
	opt := func(p *Parser) {
		p.highlightKindMap[strings.ToLower(kind)] = color
	}

	return opt
}

// Parse parses the given sequence of [resource.Resource] items in order to
// generate a directed [graph.Graph].
func (p *Parser) Parse(resources []*resource.Resource) (graph.Graph[string], error) {
	g := graph.New[string](graph.KindDirected)

	for _, r := range resources {
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

	return g, nil
}

// applyHighlights applies the highlight styles to the [graph.Vertex] u for
// [resource.Resource] r.
func (p *Parser) applyHighlights(u *graph.Vertex[string], r *resource.Resource) {
	kind := strings.ToLower(r.GetKind())
	color, ok := p.highlightKindMap[kind]
	if ok {
		u.DotAttributes["color"] = color
		u.DotAttributes["fillcolor"] = color
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
	switch {
	case origin.ConfiguredIn != "":
		// Generator or transformer created resource
		return origin.ConfiguredIn
	case origin.Repo != "":
		// Remote resource
		if origin.Ref != "" {
			return fmt.Sprintf("%s?ref=%s", origin.Path, origin.Ref)
		}
		return origin.Path
	default:
		// Local resource
		return origin.Path
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
		return origin.Repo
	default:
		// Local resource
		return ""
	}
}
