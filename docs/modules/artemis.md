# Apache ActiveMQ Artemis

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.23.0"><span class="tc-version">:material-tag: v0.23.0</span></a>

## Introduction

The Testcontainers module for Artemis.

## Adding this module to your project dependencies

Please run the following command to add the Artemis module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/artemis
```

## Usage example

<!--codeinclude-->
[Creating to an Artemis container](../../modules/artemis/examples_test.go) inside_block:runArtemisContainer
<!--/codeinclude-->

<!--codeinclude-->
[Connecting to an Artemis container](../../modules/artemis/examples_test.go) inside_block:connectToArtemisContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Artemis module exposes one entrypoint function to create the Artemis container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Artemis container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "apache/activemq-artemis:2.30.0")`.

{% include "../features/common_functional_options.md" %}

#### Credentials

If you need to change the default admin credentials (i.e. `artemis:artemis`) use `WithCredentials`.

<!--codeinclude-->
[With credentials](../../modules/artemis/artemis_test.go) inside_block:withCredentials
<!--/codeinclude-->

#### Enable Anonymous login

If you need to enable anonymous logins (which are disabled by default) use `WithAnonymousLogin`.

<!--codeinclude-->
[With Anonymous Login](../../modules/artemis/artemis_test.go) inside_block:withAnonymousLogin
<!--/codeinclude-->

#### Custom Arguments

If you need to pass custom arguments to the `artemis create` command, use `WithExtraArgs`.
The default is `--http-host 0.0.0.0 --relax-jolokia`.
Setting this value will override the default.

!!!info
    Please see the documentation on `artemis create` for the available options here: [https://activemq.apache.org/components/artemis/documentation/latest/using-server.html#options](https://activemq.apache.org/components/artemis/documentation/latest/using-server.html#options)

<!--codeinclude-->
[With Extra Arguments](../../modules/artemis/artemis_test.go) inside_block:withExtraArgs
<!--/codeinclude-->

### Container Methods

The Artemis container exposes the following methods:

#### User

User returns the administrator username.

<!--codeinclude-->
[Retrieving the Administrator User](../../modules/artemis/examples_test.go) inside_block:containerUser
<!--/codeinclude-->

#### Password

Password returns the administrator password.

<!--codeinclude-->
[Retrieving the Administrator Password](../../modules/artemis/examples_test.go) inside_block:containerPassword
<!--/codeinclude-->

#### BrokerEndpoint

BrokerEndpoint returns the host:port for the combined protocols endpoint.

<!--codeinclude-->
[Get broker endpoint](../../modules/artemis/artemis_test.go) inside_block:brokerEndpoint
<!--/codeinclude-->

#### ConsoleURL

ConsoleURL returns the URL for the management console.

<!--codeinclude-->
[Get console URL](../../modules/artemis/artemis_test.go) inside_block:consoleURL
<!--/codeinclude-->
