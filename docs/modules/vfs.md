# vfs

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for vfs.

## Adding this module to your project dependencies

Please run the following command to add the vfs module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/vfs
```

## Usage example

<!--codeinclude-->
[Creating a vfs container](../../modules/vfs/examples_test.go) inside_block:runvfsContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The vfs module exposes one entrypoint function to create the vfs container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*vfsContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the vfs container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different vfs Docker image, you can set a valid Docker image as the second argument in the `Run` function.
E.g. `Run(context.Background(), "edapt-docker-dev.artifactory.metro.ad.selinc.com/vfs:1.0.6-24289.081e1b3.develop")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The vfs container exposes the following methods:
