# Trino Gateway

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Trino Gateway.

[Trino Gateway](https://trinodb.github.io/trino-gateway/) is a load balancer, proxy server, and configurable routing gateway for multiple [Trino](https://trino.io/) clusters. Trino clients connect to the gateway as if it were a Trino coordinator, and the gateway routes their queries to the registered Trino clusters, called backends.

## Adding this module to your project dependencies

Please run the following command to add the Trino Gateway module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/trinogateway
```

## Usage example

<!--codeinclude-->
[Creating a Trino Gateway container](../../modules/trinogateway/examples_test.go) inside_block:runTrinoGatewayContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Trino Gateway module exposes one entrypoint function to create the Trino Gateway container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "trinodb/trino-gateway:22")`.

### Data store

Trino Gateway requires a relational database as data store, where it keeps the registered backends and the query history.
By default, the module starts a PostgreSQL container, using the `postgres:17-alpine` image, and attaches it together with the gateway to a dedicated Docker network.
Both containers and the network are removed when the gateway container is terminated.

Use the [`WithPostgres`](#withpostgres) option to customise the PostgreSQL container, or the [`WithDataStore`](#withdatastore) option to use an external database instead.

### Wait strategy

The module waits for the Trino Gateway to be ready by polling two HTTP endpoints on port `8080/tcp`: `/trino-gateway/readyz`, which returns `200` once the gateway's cluster monitor has completed its first run, and `/api/public/backends`, which verifies that the data store is reachable and migrated. The default startup timeout is **3 minutes**.

### Container Options

When starting the Trino Gateway container, you can pass options in a variadic way to configure it.

#### WithPostgres

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithPostgres(img string, opts ...testcontainers.ContainerCustomizer)` option to set the image and the options of the PostgreSQL container used as data store, e.g. `trinogateway.WithPostgres("postgres:16-alpine", postgres.WithPassword("secret"))`.

The option is ignored when an external data store is configured with `WithDataStore`.

#### WithDataStore

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithDataStore(jdbcURL, user, password, driver string)` option to configure Trino Gateway to use an external PostgreSQL, MySQL or Oracle database as data store, instead of starting a PostgreSQL container. The JDBC URL must be reachable from the gateway container, so the database typically shares a Docker network with it, and the driver is the JDBC driver class, e.g. `org.postgresql.Driver`, `com.mysql.cj.jdbc.Driver` or `oracle.jdbc.OracleDriver`.

```golang
gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22",
    trinogateway.WithDataStore("jdbc:postgresql://db:5432/gateway", "gateway", "secret", "org.postgresql.Driver"),
    network.WithNetwork([]string{"gateway"}, nw),
)
```

#### WithConfig

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithConfig(r io.Reader)` option to replace the default configuration file of Trino Gateway with your own, e.g. to configure [routing rules](https://trinodb.github.io/trino-gateway/routing-rules/) or authentication. The file is copied to `/etc/trino-gateway/config.yaml` in the container.

The default configuration file is the following:

<!--codeinclude-->
[Default configuration](../../modules/trinogateway/config.yaml)
<!--/codeinclude-->

The data store settings are resolved from environment variables set by the module on the gateway container, using the `${ENV:NAME}` syntax supported by Trino Gateway. A custom configuration file can reference the same variables: `DATASTORE_JDBC_URL`, `DATASTORE_USER`, `DATASTORE_PASSWORD` and `DATASTORE_DRIVER`.

!!!warning
    The gateway checks the health of the backends every `monitor.taskDelay`, and only routes queries to healthy backends. The default configuration sets it to `1s` so that a backend becomes routable shortly after it is added; Trino Gateway's own default is one minute.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Trino Gateway container exposes the following methods:

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the HTTP connection string for the Trino Gateway, e.g. `http://localhost:8080`. Trino clients, such as the [Trino Go client](https://github.com/trinodb/trino-go-client), connect to this URL as if it were a Trino coordinator.

```golang
connStr, err := gatewayContainer.ConnectionString(ctx)
```

#### Network

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the Docker network the gateway container is attached to. Trino backends must be attached to this network, so that the gateway can reach them by their network alias.

#### PostgresContainer

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the PostgreSQL container used as data store, or `nil` if an external data store was configured with `WithDataStore`.

#### Backends API

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The following methods wrap the [Trino Gateway REST API](https://trinodb.github.io/trino-gateway/gateway-api/) to manage the Trino clusters registered in the gateway:

- `AddBackend(ctx, backend Backend) error` registers a Trino cluster. The `ProxyTo` URL must be reachable from the gateway container, e.g. `http://trino:8080` for a Trino container with the `trino` network alias. The backend is registered as active, in the `adhoc` routing group unless `RoutingGroup` is set.
- `DeleteBackend(ctx, name string) error` removes a Trino cluster.
- `ActivateBackend(ctx, name string) error` and `DeactivateBackend(ctx, name string) error` control whether queries are routed to a Trino cluster.
- `Backends(ctx) ([]Backend, error)` and `ActiveBackends(ctx) ([]Backend, error)` list the registered and the active Trino clusters.
- `BackendState(ctx, name string) (BackendState, error)` returns the state of a Trino cluster as observed by the gateway's cluster monitor, including its `TrinoStatus`: `HEALTHY`, `UNHEALTHY`, `PENDING` or `UNKNOWN`.
- `WaitForBackendHealthy(ctx, name string) error` polls the state of a Trino cluster until the gateway reports it as healthy, or the context is done. Queries are routed to a backend only once it is healthy.

### Examples

#### Routing queries to a Trino backend

The following example starts a Trino container as backend of the gateway, and runs a query through the gateway using the [Trino REST API](https://trino.io/docs/current/develop/client-protocol.html). Any Trino client, such as the Trino CLI, the JDBC driver or the Go client, works the same way when pointed to the gateway's connection string.

The Trino container is attached to the gateway network, and must be configured with `http-server.process-forwarded=true`, so that the URLs it returns for the next results of a query point to the gateway instead of the Trino container. See the [Trino Gateway documentation](https://trinodb.github.io/trino-gateway/installation/#trino-configuration) for more details.

<!--codeinclude-->
[Run the Trino backend](../../modules/trinogateway/examples_test.go) inside_block:runTrinoBackend
[Register the backend](../../modules/trinogateway/examples_test.go) inside_block:addBackend
[Query through the gateway](../../modules/trinogateway/examples_test.go) inside_block:queryThroughGateway
[Run a query with the Trino REST API](../../modules/trinogateway/examples_test.go) inside_block:runQuery
<!--/codeinclude-->
