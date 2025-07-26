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
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SolaceContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "solace/solace-pubsub-standard:latest")`.

### Container Options

When starting the Solace Pubsub+ container, you can pass options in a variadic way to configure it.

#### Available Options

- `WithCredentials(username, password string)` - sets the client credentials for authentication
- `WithVpn(vpn string)` - sets the VPN name (defaults to "default")
- `WithQueue(queueName, topic string)` - subscribes a given topic to a queue (for SMF/AMQP testing)
- `WithServices(srv ...Service)` - configures which Solace services to expose with their wait strategies (preferred method)
- `WithEnv(env map[string]string)` - allows adding or overriding environment variables
- `WithShmSize(size int64)` - sets the shared memory size (defaults to 1 GiB)

#### WithServices Option

The `WithServices` option is the recommended way to configure which Solace services should be exposed and made available in your container. This option automatically handles port exposure and sets up wait strategies for each specified service.

Available services:
- `ServiceAMQP` - AMQP service (port 5672)
- `ServiceMQTT` - MQTT service (port 1883)  
- `ServiceREST` - REST service (port 9000)
- `ServiceManagement` - Management service (port 8080)
- `ServiceSMF` - SMF service (port 55555)
- `ServiceSMFSSL` - SMF SSL service (port 55443)

By default, when no `WithServices` option is specified, the container will expose AMQP, SMF, REST, and MQTT services.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Solace Pubsub+ container exposes the following methods:

#### BrokerURLFor

`BrokerURLFor(ctx context.Context, service Service) (string, error)` - returns the connection URL for a given Solace service.

This method allows you to retrieve the connection URL for specific Solace services. The available services are:

- `ServiceAMQP` - AMQP service (port 5672, protocol: amqp)
- `ServiceMQTT` - MQTT service (port 1883, protocol: tcp)  
- `ServiceREST` - REST service (port 9000, protocol: http)
- `ServiceManagement` - Management service (port 8080, protocol: http)
- `ServiceSMF` - SMF service (port 55555, protocol: tcp)
- `ServiceSMFSSL` - SMF SSL service (port 55443, protocol: tcps)

```go
// Get the AMQP connection URL
amqpURL, err := container.BrokerURLFor(ctx, solace.ServiceAMQP)
if err != nil {
    log.Fatal(err)
}
// amqpURL will be something like: amqp://localhost:32768

// Get the management URL
mgmtURL, err := container.BrokerURLFor(ctx, solace.ServiceManagement)
if err != nil {
    log.Fatal(err)
}
// mgmtURL will be something like: http://localhost:32769
```

### Container Properties

The Solace Pubsub+ container also exposes these public methods:

#### Username

`Username() string` - returns the configured username for authentication

#### Password

`Password() string` - returns the configured password for authentication

#### VPN

`VPN() string` - returns the configured VPN name

