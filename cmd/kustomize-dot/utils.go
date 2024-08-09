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
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/dnaeon/kustomize-dot/pkg/parser"
	"github.com/urfave/cli/v2"
)

// errUnsupportedLayout is returned when the app was called with invalid layout
// direction.
var errUnsupportedLayout = errors.New("unsupported graph layout")

// errInvalidKV is an error which is returned when attempting to parse an
// invalid key/value pair.
var errInvalidKV = errors.New("invalid key/value pair")

// kvSeparator is the separator used to parse key/value pairs from a string,
// e.g. foo=bar, bar=baz, etc.
const kvSeparator = "="

// getLayoutDirection returns the graph layout direction from the CLI context
func getLayoutDirection(ctx *cli.Context) (parser.LayoutDirection, error) {
	supportedLayouts := []parser.LayoutDirection{
		parser.LayoutDirectionBT,
		parser.LayoutDirectionTB,
		parser.LayoutDirectionLR,
		parser.LayoutDirectionRL,
	}

	layout := parser.LayoutDirection(ctx.String("layout"))
	if !slices.Contains(supportedLayouts, layout) {
		return parser.LayoutDirection(""), fmt.Errorf("%w: %s", errUnsupportedLayout, layout)
	}

	return layout, nil
}

// kv represents a key/value pair.
type kv struct {
	key string
	val string
}

// parseKV parses a key/value pair from a string. The key/value pair is expected
// to be in the form of foo=bar, bar=baz, etc.
func parseKV(val string) (*kv, error) {
	parts := strings.Split(val, kvSeparator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: %s", errInvalidKV, val)
	}
	pair := &kv{key: parts[0], val: parts[1]}

	return pair, nil
}
