# Smocker

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Smocker.

## Adding this module to your project dependencies

Please run the following command to add the Smocker module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/smocker
```

## Usage example

<!--codeinclude-->
[Creating a Smocker container](../../modules/smocker/examples_test.go) inside_block:runSmockerContainer
<!--/codeinclude-->

## Module reference

The Smocker module exposes one entrypoint function to create the Smocker container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SmockerContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Smocker container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Smocker Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Smocker. E.g. `testcontainers.WithImage("thiht/smocker:0.18.5")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The Smocker container exposes the following methods:

#### MockURL

This method returns the Mock URL to connect to the Smocker container, using the default `8080` port.
This URL is the one that you will use to connect your application instead of the real service.

<!--codeinclude-->
[Mock URL](../../modules/smocker/smocker_test.go) inside_block:mockURL
<!--/codeinclude-->

#### ApiURL

This method returns the API URL to connect to the Smocker container, using the default `8081` port.
This URL is the one that you will use to configure your request/response mocks.

<!--codeinclude-->
[API URL](../../modules/smocker/smocker_test.go) inside_block:apiURL
<!--/codeinclude-->
