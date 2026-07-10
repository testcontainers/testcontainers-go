# ActiveMQ

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Apache ActiveMQ Classic. This module is distinct from the [Artemis](./artemis.md) module, which targets Apache ActiveMQ Artemis (the next-generation broker).

## Adding this module to your project dependencies

Please run the following command to add the ActiveMQ module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/activemq
```

## Usage example

<!--codeinclude-->
[Creating a ActiveMQ container](../../modules/activemq/examples_test.go) inside_block:runActiveMQContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The ActiveMQ module exposes one entrypoint function to create the ActiveMQ container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "apache/activemq-classic:5.18")`.

### Container Options

When starting the ActiveMQ container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

#### WithAdminCredentials

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the username and password for the ActiveMQ web console administrator. The credentials are
propagated via the `ACTIVEMQ_WEB_ADMIN_NAME` and `ACTIVEMQ_WEB_ADMIN_PASSWORD` environment
variables. The wait strategy is also updated to authenticate with the provided credentials.

The default credentials are `admin`/`admin`.

<!--codeinclude-->
[With admin credentials](../../modules/activemq/activemq_test.go) inside_block:withAdminCredentials
<!--/codeinclude-->

### Container Methods

The ActiveMQ container exposes the following methods:

#### AdminUser

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `AdminUser()` method returns the administrator username used to access the web console.

<!--codeinclude-->
[Get admin user](../../modules/activemq/examples_test.go) inside_block:adminUser
<!--/codeinclude-->

#### AdminPassword

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `AdminPassword()` method returns the administrator password used to access the web console.

<!--codeinclude-->
[Get admin password](../../modules/activemq/examples_test.go) inside_block:adminPassword
<!--/codeinclude-->

#### BrokerURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `BrokerURL(ctx)` method returns the OpenWire broker URL as a string with the
format `tcp://<host>:<port>`, using the mapped `61616/tcp` port.

<!--codeinclude-->
[Get broker URL](../../modules/activemq/activemq_test.go) inside_block:brokerURL
<!--/codeinclude-->

#### WebConsoleURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The `WebConsoleURL(ctx)` method returns the HTTP URL of the ActiveMQ web console as a string
with the format `http://<host>:<port>`, using the mapped `8161/tcp` port.

<!--codeinclude-->
[Get web console URL](../../modules/activemq/activemq_test.go) inside_block:webConsoleURL
<!--/codeinclude-->
