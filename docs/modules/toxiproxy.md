# Toxiproxy

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

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

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

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

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "shopify/toxiproxy:2.12.0")`.

### Container Options

When starting the Toxiproxy container, you can pass options in a variadic way to configure it.

#### WithProxy

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

The `WithProxy` option allows you to specify a proxy to be created on the Toxiproxy container.
This option allocates a random port on the host and exposes it to the Toxiproxy container, allowing
you to create a unique proxy for a given service, starting from the `8666/tcp` port.

```golang
func WithProxy(name string, upstream string) Option
```

If this option is used in combination with the `WithConfigFile` option, the proxy defined in this option
is added to the proxies defined in the config file.

!!!info
    If you add proxies in a programmatic manner using the Toxiproxy client, then you need to manually
    add exposed ports in the Toxiproxy container.

#### WithConfigFile

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

The `WithConfigFile` option allows you to specify a config file for the Toxiproxy container, in the form of an `io.Reader` representing
the JSON file with the Toxiproxy configuration, in the valid format of the Toxiproxy configuration file.

<!--codeinclude-->
[Configuration file](../../modules/toxiproxy/testdata/toxiproxy.json)
<!--/codeinclude-->

```golang
func WithConfigFile(r io.Reader) testcontainers.CustomizeRequestOption
```

If this option is used in combination with the `WithProxy` option, the proxies defined in this option
are added to the proxies defined with the `WithProxy` option.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Toxiproxy container exposes the following methods:

#### ProxiedEndpoint

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

The `ProxiedEndpoint` method returns the host and port of the proxy for a given port. It's used to create new connections to the proxied service, and it returns an error in case the port has no proxy.

```golang
func (c *Container) ProxiedEndpoint(port int) (string, string, error)
```

<!--codeinclude-->
[Get Proxied Endpoint](../../modules/toxiproxy/examples_test.go) inside_block:getProxiedEndpoint
[Read Proxied Endpoint](../../modules/toxiproxy/examples_test.go) inside_block:readProxiedEndpoint
<!--/codeinclude-->

The above examples show how to get the proxied endpoint and use it to create a new connection to the proxied service, in this case a Redis client.

#### URI

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

The `URI` method returns the URI of the Toxiproxy container, used to create a new Toxiproxy client.

```golang
func (c *Container) URI() string
```

<!--codeinclude-->
[Creating a Toxiproxy client](../../modules/toxiproxy/examples_test.go) inside_block:createToxiproxyClient
<!--/codeinclude-->

- the `toxiproxy` package comes from the `github.com/Shopify/toxiproxy/v2/client` package.
- the `toxiproxyContainer` variable has been created by the `Run` function.

### Examples

#### Programmatically create a proxy

<!--codeinclude-->
[Expose port manually](../../modules/toxiproxy/examples_test.go) inside_block:defineContainerExposingPort
[Creating a proxy](../../modules/toxiproxy/examples_test.go) inside_block:createProxy
[Creating a Redis client](../../modules/toxiproxy/examples_test.go) inside_block:createRedisClient
[Adding a latency toxic](../../modules/toxiproxy/examples_test.go) inside_block:addLatencyToxic

<!--/codeinclude-->