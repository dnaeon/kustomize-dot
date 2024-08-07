# kustomize-dot

[![Build Status](https://github.com/dnaeon/kustomize-dot/actions/workflows/test.yaml/badge.svg)](https://github.com/dnaeon/kustomize-dot/actions/workflows/test.yaml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/dnaeon/kustomize-dot.svg)](https://pkg.go.dev/github.com/dnaeon/kustomize-dot)
[![Go Report Card](https://goreportcard.com/badge/github.com/dnaeon/kustomize-dot)](https://goreportcard.com/report/github.com/dnaeon/kustomize-dot)
[![codecov](https://codecov.io/gh/dnaeon/kustomize-dot/branch/master/graph/badge.svg)](https://codecov.io/gh/dnaeon/kustomize-dot)

`kustomize-dot` is a kustomize plugin which generates a dependency graph of the
Kubernetes resources produced by `kustomize build`.

## Requirements

* Go version 1.21.x or later
* Docker for local development

## Installation

TODO: instructions on how to build it via makefile target and `go install`
TODO: instructions on how to install the packages

## Usage

TODO: instructions on how to use it as a KRM Function Plugin
TODO: instructions on how to use it as a standalone CLI tool

## Tests

Run the tests.

``` shell
make test
```

Run test coverage.

``` shell
make test-cover
```

## Contributing

`kustomize-dot` is hosted on
[Github](https://github.com/dnaeon/kustomize-dot). Please contribute by
reporting issues, suggesting features or by sending patches using pull requests.

## License

`kustomize-dot` is Open Source and licensed under the [BSD
License](http://opensource.org/licenses/BSD-2-Clause).
