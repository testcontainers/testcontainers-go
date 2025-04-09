# Toxiproxy

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Toxiproxy.

## Adding this module to your project dependencies

Please run the following command to add the Toxiproxy module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/toxiproxy
```

## Usage example

<!--codeinclude-->
[Creating a Toxiproxy container](../../modules/toxiproxy/examples_test.go) inside_block:runToxiproxyContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Toxiproxy module exposes one entrypoint function to create the Toxiproxy container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports

The Toxiproxy container exposes the following ports:

- `8474/tcp`, the Toxiproxy control port, exported as `toxiproxy.ControlPort`.

### Container Options

When starting the Toxiproxy container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "shopify/toxiproxy:2.12.0")`.

{% include "../features/common_functional_options.md" %}

#### WithPortRange

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `WithPortRange` option allows you to specify the number of ports to expose on the Toxiproxy container.
This option allocates a range of ports on the host and exposes them to the Toxiproxy container, allowing
you to create a unique proxy for each container. The default port range is `31`.

```golang
func WithPortRange(portRange int) Option
```

#### WithConfigFile

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `WithConfigFile` option allows you to specify a config file for the Toxiproxy container, in the form of an `io.Reader` representing
the JSON file with the Toxiproxy configuration.

!!! warning
    The config file is not validated by the Toxiproxy container.

```golang
func WithConfigFile(r io.Reader) testcontainers.CustomizeRequestOption
```

### Container Methods

The Toxiproxy container exposes the following methods:

#### URI

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `URI` method returns the URI of the Toxiproxy container, used to create a new Toxiproxy client.

```golang
func (c *ToxiproxyContainer) URI() string
```

<!--codeinclude-->
[Creating a Toxiproxy client](../../modules/toxiproxy/examples_test.go) inside_block:createToxiproxyClient
<!--/codeinclude-->

- the `toxiproxy` package comes from the `github.com/Shopify/toxiproxy/v2/client` package.
- the `toxiproxyContainer` variable has been created by the `Run` function.

### Examples

#### Adding a latency toxic to the proxy

<!--codeinclude-->
[Creating a proxy](../../modules/toxiproxy/examples_test.go) inside_block:createProxy
[Creating a Redis client](../../modules/toxiproxy/examples_test.go) inside_block:createRedisClient
[Adding a latency toxic](../../modules/toxiproxy/examples_test.go) inside_block:addLatencyToxic

<!--/codeinclude-->