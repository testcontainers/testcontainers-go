# Mosquitto

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Eclipse Mosquitto](https://mosquitto.org/), a lightweight open-source MQTT broker that implements the MQTT protocol versions 5, 3.1.1, and 3.1. Commonly used for IoT and messaging workloads.

## Adding this module to your project dependencies

Please run the following command to add the Mosquitto module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mosquitto
```

## Usage example

<!--codeinclude-->
[Creating a Mosquitto container](../../modules/mosquitto/examples_test.go) inside_block:runMosquittoContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Mosquitto module exposes one entrypoint function to create the Mosquitto container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "eclipse-mosquitto:2")`.

!!!info
    The `eclipse-mosquitto` image requires a configuration file with at least a `listener 1883` directive to accept connections. The module automatically injects a minimal default configuration that enables anonymous connections on port 1883 when no custom configuration is provided.

### Container Options

When starting the Mosquitto container, you can pass options in a variadic way to configure it.

#### WithConfigFile

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Mounts a custom `mosquitto.conf` file, replacing the module's default embedded configuration. The provided file must contain at least a `listener 1883` directive.

```golang
mosquittoContainer, err := mosquitto.Run(ctx, "eclipse-mosquitto:2",
    mosquitto.WithConfigFile("/path/to/mosquitto.conf"),
)
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Mosquitto container exposes the following methods:

#### BrokerURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the MQTT broker URL (`mqtt://host:port`) for connecting MQTT clients to the broker.

```golang
brokerURL, err := mosquittoContainer.BrokerURL(ctx)
```
