# Consul

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for Consul.

## Adding this module to your project dependencies

Please run the following command to add the Consul module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/consul
```

## Usage example

<!--codeinclude-->
[Creating a Consul container](../../modules/consul/examples_test.go) inside_block:runConsulContainer
<!--/codeinclude-->

## Module reference

The Consul module exposes one entrypoint function to create the Consul container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ConsulContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Consul container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Consul Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Consul. E.g. `testcontainers.WithImage("docker.io/hashicorp/consul:1.15")`.

{% include "../features/common_functional_options.md" %}

#### Configuration File
If you need to customize the behavior for the deployed node you can use either `WithConfigString(config string)` or `WithConfigFile(configPath string)`.
The configuration has to be in JSON format and will be loaded at the node startup.

### Container Methods

The Consul container exposes the following method:

#### ApiEndpoint
This method returns the connection string to connect to the Consul container API, using the default `8500` port.

<!--codeinclude-->
[Using ApiEndpoint with the Consul client](../../modules/consul/examples_test.go) inside_block:connectConsul
<!--/codeinclude-->
