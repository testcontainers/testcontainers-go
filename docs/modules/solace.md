# Solace Pubsub+

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Solace Pubsub+.

## Adding this module to your project dependencies

Please run the following command to add the Solace Pubsub+ module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/solace
```

## Usage example

<!--codeinclude-->
[Creating a Solace Pubsub+ container](../../modules/solace/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Solace Pubsub+ module exposes one entrypoint function to create the Solace Pubsub+ container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "solace-pubsub-standard:latest")`.

### Container Options

When starting the Solace Pubsub+ container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Solace Pubsub+ container exposes the following methods:

#### BrokerURLFor

`BrokerURLFor(service Service) (string, error)` - returns the connection URL for a given Solace service.

This method allows you to retrieve the connection URL for specific Solace services. The available services are:

- `ServiceAMQP` - AMQP service (port 5672)
- `ServiceMQTT` - MQTT service (port 1883)  
- `ServiceREST` - REST service (port 9000)
- `ServiceManagement` - Management service (port 8080)
- `ServiceSMF` - SMF service (port 55555)
- `ServiceSMFSSL` - SMF SSL service (port 55443)

```go
// Get the AMQP connection URL
amqpURL, err := container.BrokerURLFor(solace.ServiceAMQP)
if err != nil {
    log.Fatal(err)
}
// amqpURL will be something like: amqp://localhost:32768

// Get the management URL
mgmtURL, err := container.BrokerURLFor(solace.ServiceManagement)
if err != nil {
    log.Fatal(err)
}
// mgmtURL will be something like: http://localhost:32769
```

#### Terminate

`Terminate() error` - terminates the Solace container.

```go
err := container.Terminate()
if err != nil {
    log.Fatal(err)
}
```

### Container Properties

The Solace Pubsub+ container also exposes these public properties:

- `Username` - the configured username for authentication
- `Password` - the configured password for authentication  
- `Vpn` - the configured VPN name
- `Container` - the underlying testcontainers Container instance

