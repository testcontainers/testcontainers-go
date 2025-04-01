# Socat

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Socat, a utility container that provides TCP port forwarding and network tunneling between services, enabling transparent communication between containers and networks.

This is particularly useful in testing scenarios where you need to simulate network connections or provide transparent access to services running in different containers.

## Adding this module to your project dependencies

Please run the following command to add the Socat module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/socat
```

## Usage example

<!--codeinclude-->
[Create a Network](../../modules/socat/examples_test.go) inside_block:createNetwork
[Create a Hello World Container](../../modules/socat/examples_test.go) inside_block:createHelloWorldContainer
[Create a Socat Container](../../modules/socat/examples_test.go) inside_block:createSocatContainer
[Read from Socat Container](../../modules/socat/examples_test.go) inside_block:readFromSocat
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Socat module exposes one entrypoint function to create the Socat container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SocatContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Socat container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "alpine/socat:1.8.0.1")`.

{% include "../features/common_functional_options.md" %}

#### WithTarget

The `WithTarget` function sets a single target for the Socat container, defined by the `Target` struct.
This struct can be built using the the following functions:

- `NewTarget(exposedPort int, host string)`: Creates a new target for the Socat container. The target's internal port is set to the same value as the exposed port.
- `NewTargetWithInternalPort(exposedPort int, internalPort int, host string)`: Creates a new target for the Socat container with an internal port. Use this function when you want to map a container to a different port than the default one.

<!--codeinclude-->
[Passing a target](../../modules/socat/examples_test.go) inside_block:createSocatContainer
<!--/codeinclude-->

In the above example, there is a `helloworld` container thatis listening on port `8080` and `8081`. Please check [the helloworld container source code](https://github.com/testcontainers/helloworld/blob/141af7909907e04b124e691d3cd6fc7c32da2207/internal/server/server.go#L26-L27) for more details.

### Container Methods

The Socat container exposes the following methods:

#### TargetURL

The `TargetURL(port int)` method returns the URL for the exposed port of a target, nil if the port is not mapped.

<!--codeinclude-->
[Read from Socat using TargetURL](../../modules/socat/examples_test.go) inside_block:readFromSocat
<!--/codeinclude-->
