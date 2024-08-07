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
// are connected via edges to their origin metadata.
type Parser struct{}

// Parse parses the given sequence of [resource.Resource] items in order to
// generate a directed [graph.Graph].
func (p *Parser) Parse(resources []*resource.Resource) (graph.Graph[string], error) {
	g := graph.New[string](graph.KindDirected)

	for _, resource := range resources {
		u := vertexNameFromResource(resource)
		g.AddVertex(u)

		origin, err := resource.GetOrigin()
		if err != nil {
			return nil, err
		}

		if origin == nil {
			continue
		}

		v := vertexNameFromOrigin(origin)
		g.AddVertex(v)

		e := g.AddEdge(u, v)
		label := edgeLabelFromOrigin(origin)
		e.DotAttributes["label"] = label
	}

	return g, nil
}

// vertexNameFromResource returns a string representing the vertex name for the
// given [resource.Resource].
func vertexNameFromResource(r *resource.Resource) string {
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
func vertexNameFromOrigin(origin *resource.Origin) string {
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
func edgeLabelFromOrigin(origin *resource.Origin) string {
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
