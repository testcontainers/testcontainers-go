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

A data store must be selected with the [`WithDataStore`](#withdatastore) option, otherwise `Run` returns an error.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "trinodb/trino-gateway:22")`.

### Data store

Trino Gateway requires a relational database as data store, where it keeps the registered backends and the query history. The database is described by a data store provider, which defines how the gateway gets the database and who owns its lifecycle. The provider must be selected explicitly with the [`WithDataStore`](#withdatastore) option, so that starting a database is always visible at the call site.

The module creates a Docker network for the gateway, and owns it. Terminating the gateway container terminates, in order: the gateway container, the resources created by the data store provider, and the network. A database owned by the caller is never terminated.

#### Managed PostgreSQL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`NewManagedPostgresProvider(img string, opts ...testcontainers.ContainerCustomizer)` returns a provider that starts a PostgreSQL container with the [postgres module](postgres.md), on the network created for the gateway, using the `postgres` network alias. The options are the options of the postgres module, e.g. `postgres.WithPassword("secret")`.

The provider owns the PostgreSQL container: the container is terminated with the gateway, and if it fails to start, the provider removes it before `Run` returns the error. A provider starts a single container, so it cannot be used by two gateways at the same time.

Use the `Container()` method of the provider to access the `*postgres.PostgresContainer`, e.g. to read its connection string. It returns `nil` before the container is started, and after it is terminated.

```golang
provider := trinogateway.NewManagedPostgresProvider("postgres:17-alpine")

gatewayContainer, err := trinogateway.Run(ctx, "trinodb/trino-gateway:22", trinogateway.WithDataStore(provider))
// handle the error and terminate the container

pgContainer := provider.Container()
```

#### External database

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`NewExternalProvider(jdbcURL, user, password, driver string)` returns a provider for a PostgreSQL, MySQL or Oracle database started by the caller, such as another container or an existing database service. The provider creates no database resources, and never terminates the database.

The caller ensures that the JDBC URL is reachable from the gateway container, e.g. attaching the gateway to the network of the database with `network.WithNetwork`. The driver is the JDBC driver class, e.g. `org.postgresql.Driver`, `com.mysql.cj.jdbc.Driver` or `oracle.jdbc.OracleDriver`.

<!--codeinclude-->
[Using a database owned by the caller](../../modules/trinogateway/examples_test.go) inside_block:runWithExternalDatabase
<!--/codeinclude-->

### Wait strategy

The module waits for the Trino Gateway to be ready by polling two HTTP endpoints on port `8080/tcp`: `/trino-gateway/readyz`, which returns `200` once the gateway's cluster monitor has completed its first run, and `/api/public/backends`, which verifies that the data store is reachable and migrated. The default startup timeout is **3 minutes**.

### Container Options

When starting the Trino Gateway container, you can pass options in a variadic way to configure it.

#### WithDataStore

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithDataStore(provider DataStoreProvider)` option to select the data store of the gateway: a [managed PostgreSQL container](#managed-postgresql), with `NewManagedPostgresProvider`, or an [external database](#external-database), with `NewExternalProvider`. This option is required.

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

This method returns the HTTP connection string for the Trino Gateway, e.g. `http://localhost:8080`. Trino clients connect to this URL as if it were a Trino coordinator.

```golang
connStr, err := gatewayContainer.ConnectionString(ctx)
```

#### Network

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the Docker network created for the gateway container, which is also used by a managed PostgreSQL data store. Trino backends must be attached to this network, so that the gateway can reach them by their network alias.

#### AddBackend

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`AddBackend(ctx, backend Backend) error` registers a Trino cluster in the gateway, using the [Trino Gateway REST API](https://trinodb.github.io/trino-gateway/gateway-api/). The `ProxyTo` URL must be reachable from the gateway container, e.g. `http://trino:8080` for a Trino container with the `trino` network alias. The backend is registered as active, in the `adhoc` routing group unless `RoutingGroup` is set. Queries are routed to it once the gateway reports it as healthy, see [`WaitForBackendHealthy`](#waitforbackendhealthy).

#### DeleteBackend

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`DeleteBackend(ctx, name string) error` removes a Trino cluster from the gateway.

#### ActivateBackend

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`ActivateBackend(ctx, name string) error` activates a Trino cluster, so that queries are routed to it.

#### DeactivateBackend

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`DeactivateBackend(ctx, name string) error` deactivates a Trino cluster, so that no queries are routed to it.

#### Backends

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`Backends(ctx) ([]Backend, error)` returns all the Trino clusters registered in the gateway.

#### ActiveBackends

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`ActiveBackends(ctx) ([]Backend, error)` returns the active Trino clusters registered in the gateway.

#### BackendState

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`BackendState(ctx, name string) (BackendState, error)` returns the state of a Trino cluster as observed by the gateway's cluster monitor, including its `TrinoStatus`: `HEALTHY`, `UNHEALTHY`, `PENDING` or `UNKNOWN`.

#### WaitForBackendHealthy

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WaitForBackendHealthy(ctx, name string) error` polls the state of a Trino cluster until the gateway reports it as healthy, or the context is done. The gateway only routes queries to healthy backends.

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
