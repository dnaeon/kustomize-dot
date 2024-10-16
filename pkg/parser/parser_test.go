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

package parser

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/dnaeon/kustomize-dot/pkg/fixtures"

	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestParseResources(t *testing.T) {
	type testCase struct {
		data          string
		desc          string
		wantResources int
		wantError     error
	}

	testCases := []testCase{
		{
			desc:          "empty data",
			data:          "",
			wantResources: 0,
			wantError:     nil,
		},
		{
			desc: "single resource data",
			data: `
apiVersion: v1
data:
  altGreeting: Good Morning!
  enableRisky: "false"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: examples/helloWorld/configMap.yaml
      repo: https://github.com/kubernetes-sigs/kustomize
      ref: v1.0.6
  labels:
    app: hello
  name: the-map
`,
			wantError:     nil,
			wantResources: 1,
		},
		{
			desc:          "hello world resources",
			data:          fixtures.HelloWorld,
			wantResources: 3,
			wantError:     nil,
		},
		{
			desc:          "kube prometheus resources",
			data:          fixtures.KubePrometheus,
			wantResources: 124,
			wantError:     nil,
		},
		{
			desc:          "bad data",
			data:          "some bad data in here",
			wantResources: 0,
			wantError:     yaml.ErrMissingMetadata,
		},
	}

	for _, tc := range testCases {
		// Parse resources from bytes
		t.Run(fmt.Sprintf("ResourcesFromBytes with %s", tc.desc), func(t *testing.T) {
			gotResources, err := ResourcesFromBytes([]byte(tc.data))
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("want %v error, got %v", tc.wantError, err)
			}

			if len(gotResources) != tc.wantResources {
				t.Fatalf("got %d resource(s), want %d", len(gotResources), tc.wantResources)
			}
		})

		t.Run(fmt.Sprintf("parser.ResourcesFromReader with %s", tc.desc), func(t *testing.T) {
			reader := strings.NewReader(tc.data)
			gotResources, err := ResourcesFromReader(reader)
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("want %v error, got %v", tc.wantError, err)
			}

			if len(gotResources) != tc.wantResources {
				t.Fatalf("got %d resource(s), want %d", len(gotResources), tc.wantResources)
			}
		})
	}
}

func TestVertexNameAndEdgeLabelFromOrigin(t *testing.T) {
	type testCase struct {
		desc           string
		wantEdgeLabel  string
		wantVertexName string
		origin         *resource.Origin
	}

	testCases := []testCase{
		{
			desc:           "local resource",
			wantVertexName: "foo.yaml",
			wantEdgeLabel:  "",
			origin: &resource.Origin{
				Path: "foo.yaml",
			},
		},
		{
			desc:           "generator / transformer created resource",
			wantVertexName: "foo",
			wantEdgeLabel:  "v1/my-generator",
			origin: &resource.Origin{
				Path:         "foo.yaml",
				ConfiguredIn: "foo",
				ConfiguredBy: yaml.ResourceIdentifier{
					TypeMeta: yaml.TypeMeta{
						APIVersion: "v1",
						Kind:       "my-generator",
					},
				},
			},
		},
		{
			desc:           "remote resource without ref",
			wantVertexName: "foo.yaml",
			wantEdgeLabel:  "github.com/dnaeon/kustomize-dot",
			origin: &resource.Origin{
				Repo: "github.com/dnaeon/kustomize-dot",
				Path: "foo.yaml",
			},
		},
		{
			desc:           "remote resource with ref",
			wantVertexName: "foo.yaml",
			wantEdgeLabel:  "github.com/dnaeon/kustomize-dot (ref v1)",
			origin: &resource.Origin{
				Repo: "github.com/dnaeon/kustomize-dot",
				Ref:  "v1",
				Path: "foo.yaml",
			},
		},
	}

	p := New()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotVertexName := p.vertexNameFromOrigin(tc.origin)
			if gotVertexName != tc.wantVertexName {
				t.Fatalf("want vertex name %q, got name %q", tc.wantVertexName, gotVertexName)
			}

			gotEdgeLabel := p.edgeLabelFromOrigin(tc.origin)
			if gotEdgeLabel != tc.wantEdgeLabel {
				t.Fatalf("want edge label %q, got label %q", tc.wantEdgeLabel, gotEdgeLabel)
			}
		})
	}
}

func TestVertexNameFromResource(t *testing.T) {
	configMap, err := NewResourceFactory().FromMapWithName(
		"kustomize-dot",
		map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]string{
				"name":      "kustomize-dot",
				"namespace": "default",
			},
			"data": map[string]string{
				"foo": "bar",
			},
		},
	)
	if err != nil {
		t.Fatal("failed to create ConfigMap resource")
	}

	namespace, err := NewResourceFactory().FromMapWithName(
		"default",
		map[string]any{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]string{
				"name": "default",
			},
		},
	)
	if err != nil {
		t.Fatal("failed to create Namespace resource")
	}

	type testCase struct {
		desc     string
		wantName string
		resource *resource.Resource
	}
	testCases := []testCase{
		{
			desc:     "namespace resource",
			wantName: "namespace/default",
			resource: namespace,
		},
		{
			desc:     "ConfigMap resource",
			wantName: "default/configmap/kustomize-dot",
			resource: configMap,
		},
	}

	p := New()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotName := p.vertexNameFromResource(tc.resource)
			if gotName != tc.wantName {
				t.Fatalf("want vertex name %q, got name %q", tc.wantName, gotName)
			}
		})
	}
}

func TestWithKeepAndWithDropOptions(t *testing.T) {
	// Our test resources
	configMap, err := NewResourceFactory().FromMapWithName(
		"kustomize-dot",
		map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]string{
				"name":      "kustomize-dot",
				"namespace": "default",
			},
			"data": map[string]string{
				"foo": "bar",
			},
		},
	)
	if err != nil {
		t.Fatal("failed to create ConfigMap resource")
	}

	namespace, err := NewResourceFactory().FromMapWithName(
		"default",
		map[string]any{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]string{
				"name": "default",
			},
		},
	)
	if err != nil {
		t.Fatal("failed to create Namespace resource")
	}

	type testCase struct {
		desc       string
		r          *resource.Resource
		shouldDrop bool
		opts       []Option
	}

	testCases := []testCase{
		{
			desc:       "empty opts - should persist resource",
			r:          configMap,
			shouldDrop: false,
			opts:       []Option{},
		},
		{
			desc:       "WithDropNamespace - should drop",
			r:          configMap,
			shouldDrop: true,
			opts:       []Option{WithDropNamespace("default")},
		},
		{
			desc:       "WithDropNamespace - should persist",
			r:          configMap,
			shouldDrop: false,
			opts:       []Option{WithDropNamespace("foobar")}, // Resource is in the default namespace
		},
		{
			desc:       "WithDropKind - should drop",
			r:          namespace,
			shouldDrop: true,
			opts:       []Option{WithDropKind("Namespace")},
		},
		{
			desc:       "WithDropKind - should persist",
			r:          configMap,
			shouldDrop: false,
			opts:       []Option{WithDropKind("Secret")}, // Resource is a ConfigMap
		},
		{
			desc:       "WithKeepKind - should drop",
			r:          configMap,
			shouldDrop: true,
			opts:       []Option{WithKeepKind("Secret")}, // Resource is a ConfigMap
		},
		{
			desc:       "WithKeepKind - should persist",
			r:          configMap,
			shouldDrop: false,
			opts:       []Option{WithKeepKind("ConfigMap")},
		},
		{
			desc:       "WithKeepNamespace - should persist",
			r:          configMap,
			shouldDrop: false,
			opts:       []Option{WithKeepNamespace("default")},
		},
		{
			desc:       "WithKeepNamespace - should drop",
			r:          configMap,
			shouldDrop: true,
			opts:       []Option{WithKeepNamespace("foobar")}, // Resource is in the default namespace
		},
		{
			desc:       "WithKeepNamespace - should persist cluster scoped resource",
			r:          namespace,
			shouldDrop: false,
			opts:       []Option{WithKeepNamespace("foobar")}, // Resource is cluster-scoped
		},
		{
			desc:       "WithKeepNamespace and WithDropKind - should drop",
			r:          configMap,
			shouldDrop: true,
			opts:       []Option{WithKeepNamespace("default"), WithDropKind("ConfigMap")},
		},
		{
			desc:       "WithKeepNamespace and WithKeepKind - should drop",
			r:          configMap,
			shouldDrop: true,
			// Resource is not a Secret, so it should be dropped
			opts: []Option{WithKeepNamespace("default"), WithKeepKind("Secret")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p := New(tc.opts...)
			gotShouldDrop := p.shouldDropResource(tc.r)
			if gotShouldDrop != tc.shouldDrop {
				t.Fatalf("shouldDrop() returned %t, expected %t", gotShouldDrop, tc.shouldDrop)
			}
		})
	}
}

func TestWithHighlightOptions(t *testing.T) {
	type testCase struct {
		desc                      string
		opts                      []Option
		wantHighlightKindMap      map[string]string
		wantHighlightNamespaceMap map[string]string
	}

	testCases := []testCase{
		{
			desc:                      "empty opts - no highlights",
			opts:                      []Option{},
			wantHighlightKindMap:      map[string]string{},
			wantHighlightNamespaceMap: map[string]string{},
		},
		{
			desc: "WithHighlightKind - multiple highlights",
			opts: []Option{
				WithHighlightKind("ConfigMap", "red"),
				WithHighlightKind("Secret", "green"),
				WithHighlightKind("Namespace", "blue"),
			},
			wantHighlightKindMap: map[string]string{
				"configmap": "red",
				"secret":    "green",
				"namespace": "blue",
			},
			wantHighlightNamespaceMap: map[string]string{},
		},
		{
			desc: "WithHighlightNamespace - multiple highlights",
			opts: []Option{
				WithHighlightNamespace("foo", "red"),
				WithHighlightNamespace("bar", "green"),
				WithHighlightNamespace("baz", "blue"),
			},
			wantHighlightKindMap: map[string]string{},
			wantHighlightNamespaceMap: map[string]string{
				"foo": "red",
				"bar": "green",
				"baz": "blue",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p := New(tc.opts...)
			if !maps.Equal(p.highlightKindMap, tc.wantHighlightKindMap) {
				t.Fatalf("want highlightKindMap %v, got %v", tc.wantHighlightKindMap, p.highlightKindMap)
			}
			if !maps.Equal(p.highlightNamespaceMap, tc.wantHighlightNamespaceMap) {
				t.Fatalf("want highlightNamespaceMap %v, got %v", tc.wantHighlightNamespaceMap, p.highlightNamespaceMap)
			}
		})
	}
}

func TestWithLayoutDirection(t *testing.T) {
	type testCase struct {
		desc                string
		opts                []Option
		wantLayoutDirection LayoutDirection
		wantLayoutStr       string
	}

	testCases := []testCase{
		{
			desc:                "no layout specified - defaults to LR",
			opts:                []Option{},
			wantLayoutDirection: LayoutDirectionLR,
			wantLayoutStr:       "LR",
		},
		{
			desc:                "WithLayoutDirection specified",
			opts:                []Option{WithLayoutDirection(LayoutDirectionBT)},
			wantLayoutDirection: LayoutDirectionBT,
			wantLayoutStr:       "BT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p := New(tc.opts...)
			if p.layoutDirection != tc.wantLayoutDirection {
				t.Fatalf("want layout direction %v, got %v", tc.wantLayoutDirection, p.layoutDirection)
			}
			if p.layoutDirection.String() != tc.wantLayoutStr {
				t.Fatalf("want layout string %v, got %v", tc.wantLayoutStr, p.layoutDirection)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type testCase struct {
		desc          string
		data          string   // data from which to get raw resources
		wantResources int      // number of resources we get from the raw data
		wantVs        int      // number of expected vertices in the graph
		wantEs        int      // number of expected edges in the graph
		opts          []Option // options with which to configure the parser
	}

	testCases := []testCase{
		{
			desc:          "empty data - no options",
			data:          "",
			wantResources: 0,
			wantVs:        0,
			wantEs:        0,
			opts:          []Option{},
		},
		{
			desc:          "hello world resources - no options",
			data:          fixtures.HelloWorld,
			wantResources: 3,
			wantVs:        6, // 3 resources + 3 origins
			wantEs:        3,
			opts:          []Option{},
		},
		{
			desc:          "hello world resources - WithDropKind",
			data:          fixtures.HelloWorld,
			wantResources: 3,
			wantVs:        4, // 2 resources + 2 origins
			wantEs:        2,
			opts:          []Option{WithDropKind("Service")},
		},
		{
			desc:          "hello world resources - WithDropNamespace",
			data:          fixtures.HelloWorld,
			wantResources: 3,
			wantVs:        0,
			wantEs:        0,
			opts:          []Option{WithDropNamespace("default")},
		},
		{
			desc:          "hello world resources - WithKeepNamespace",
			data:          fixtures.HelloWorld,
			wantResources: 3,
			wantVs:        0,
			wantEs:        0,
			opts:          []Option{WithKeepNamespace("foobar")}, // Resources are from default namespace
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			reader := strings.NewReader(tc.data)
			gotResources, err := ResourcesFromReader(reader)
			if err != nil {
				t.Fatalf("parsing resources failed: %s", err)
			}

			if len(gotResources) != tc.wantResources {
				t.Fatalf("got %d resource(s), want %d", len(gotResources), tc.wantResources)
			}

			p := New(tc.opts...)
			g, err := p.Parse(gotResources)
			if err != nil {
				t.Fatalf("failed to parse resources as graph: %s", err)
			}

			gotVs := g.GetVertices()
			if tc.wantVs != len(gotVs) {
				t.Fatalf("want |V|=%d, got |V|=%d", tc.wantVs, len(gotVs))
			}

			gotEs := g.GetEdges()
			if tc.wantEs != len(gotEs) {
				t.Fatalf("want |E|=%d, got |E|=%d", tc.wantEs, len(gotEs))
			}
		})
	}
}
