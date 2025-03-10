# NATS

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

## Introduction

The Testcontainers module for NATS.

## Adding this module to your project dependencies

Please run the following command to add the NATS module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/nats
```

## Usage example

<!--codeinclude-->
[Creating a NATS container](../../modules/nats/examples_test.go) inside_block:runNATSContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The NATS module exposes one entrypoint function to create the NATS container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*NATSContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the NATS container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "nats:2.9")`.

{% include "../features/common_functional_options.md" %}

#### Set username and password

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

If you need to set different credentials, you can use `WithUsername` and `WithPassword`
options. By default, the username, the password are not set. To establish the connection with the NATS container:

<!--codeinclude-->
[Connect using the credentials](../../modules/nats/examples_test.go) inside_block:natsConnect
<!--/codeinclude-->

#### Cmd Arguments

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

It's possible to pass extra arguments to the NATS container using the `testcontainers.WithArgument` option. E.g. `nats.WithArgument("cluster_name", "c1")`.
These arguments are passed to the NATS server when it starts, as part of the command line arguments of the entrypoint.

!!! note
    Arguments do not need to be prefixed with `--`: the NATS container will add them automatically.

<!--codeinclude-->
[Passing arguments](../../modules/nats/examples_test.go) inside_block:withArguments
<!--/codeinclude-->

#### Custom configuration file

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.35.0"><span class="tc-version">:material-tag: v0.35.0</span></a>

It's possible to pass a custom config file to NATS container using `nats.WithConfigFile(strings.NewReader(config))`. The content of `io.Reader` is passed as a `-config /etc/nats.conf` arguments to an entrypoint. 

!!! note
    Changing the connectivity (listen address or ports) can break the container setup. So configuration must be done with care.

### Container Methods

The NATS container exposes the following methods:

#### ConnectionString

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.24.0"><span class="tc-version">:material-tag: v0.24.0</span></a>

This method returns the connection string to connect to the NATS container, using the default `4222` port.

<!--codeinclude-->
[Get connection string](../../modules/nats/nats_test.go) inside_block:connectionString
<!--/codeinclude-->

#### MustConnectionString

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

Exactly like `ConnectionString`, but it panics if an error occurs, returning just a string.

## Examples

### NATS Cluster

<!--codeinclude-->
[NATS Cluster](../../modules/nats/examples_test.go) inside_block:cluster
<!--/codeinclude-->
