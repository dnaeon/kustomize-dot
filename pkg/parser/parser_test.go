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

func TestEdgeLabelFromOrigin(t *testing.T) {
	type testCase struct {
		desc      string
		wantLabel string
		origin    *resource.Origin
	}

	testCases := []testCase{
		{
			desc:      "local resource",
			wantLabel: "",
			origin: &resource.Origin{
				Path: "foo.yaml",
			},
		},
		{
			desc:      "generator / transformer created resource",
			wantLabel: "v1/my-generator",
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
			desc:      "remote resource without ref",
			wantLabel: "github.com/dnaeon/kustomize-dot",
			origin: &resource.Origin{
				Repo: "github.com/dnaeon/kustomize-dot",
				Path: "foo.yaml",
			},
		},
		{
			desc:      "remote resource with ref",
			wantLabel: "github.com/dnaeon/kustomize-dot (ref v1)",
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
			gotLabel := p.edgeLabelFromOrigin(tc.origin)
			if gotLabel != tc.wantLabel {
				t.Fatalf("want edge label %q, got label %q", tc.wantLabel, gotLabel)
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
