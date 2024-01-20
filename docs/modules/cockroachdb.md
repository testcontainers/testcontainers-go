# CockroachDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for CockroachDB.

## Adding this module to your project dependencies

Please run the following command to add the CockroachDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/cockroachdb
```

## Usage example

<!--codeinclude-->
[Creating a CockroachDB container](../../modules/cockroachdb/examples_test.go) inside_block:runCockroachDBContainer
<!--/codeinclude-->

## Module reference

The CockroachDB module exposes one entrypoint function to create the CockroachDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the CockroachDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different CockroachDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for CockroachDB. E.g. `testcontainers.WithImage("cockroachdb/cockroach:latest-v23.1")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The CockroachDB container exposes the following methods:
