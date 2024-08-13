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

package parser_test

import (
	"errors"
	"testing"

	"github.com/dnaeon/kustomize-dot/pkg/parser"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestResourcesFromBytes(t *testing.T) {
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
			desc:          "bad data",
			data:          "some bad data in here",
			wantResources: 0,
			wantError:     yaml.ErrMissingMetadata,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotResources, err := parser.ResourcesFromBytes([]byte(tc.data))
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("want %v error, got %v", tc.wantError, err)
			}

			if len(gotResources) != tc.wantResources {
				t.Fatalf("got %d resource(s), want %d", len(gotResources), tc.wantResources)
			}
		})
	}
}

func TestResourcesFromRNodes(t *testing.T) {
}

func TestResourcesFromPath(t *testing.T) {
}

func TestResourcesFromReader(t *testing.T) {
}
