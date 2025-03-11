# ClickHouse

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.23.0"><span class="tc-version">:material-tag: v0.23.0</span></a>

## Introduction

The Testcontainers module for ClickHouse.

## Adding this module to your project dependencies

Please run the following command to add the ClickHouse module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/clickhouse
```

## Usage example

<!--codeinclude-->
[Test for a ClickHouse container](../../modules/clickhouse/examples_test.go) inside_block:runClickHouseContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The ClickHouse module exposes one entrypoint function to create the ClickHouse container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports

Here you can find the list with the default ports used by the ClickHouse container.

<!--codeinclude-->
[Container Ports](../../modules/clickhouse/clickhouse.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the ClickHouse container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "clickhouse/clickhouse-server:23.3.8.21-alpine")`.

{% include "../features/common_functional_options.md" %}

#### Set username, password and database name

If you need to set a different database, and its credentials, you can use `WithUsername`, `WithPassword`, `WithDatabase`
options. E.g. `WithUsername("user")`, `WithPassword("password")`, `WithDatabase("db")`.

!!!info
    The default values for the username is `default`, for password is `clickhouse` and for the default database name is `clickhouse`.

#### Init Scripts

If you would like to do additional initialization in the ClickHouse container, add one or more `*.sql`, `*.sql.gz`, or `*.sh` scripts to the container request.
Those files will be copied after the container is created but before it's started under `/docker-entrypoint-initdb.d`. According to ClickHouse Docker image,
it will run any `*.sql` files, run any executable `*.sh` scripts, and source any non-executable `*.sh` scripts found in that directory to do further
initialization before starting the service.

<!--codeinclude-->
[Include init scripts](../../modules/clickhouse/clickhouse_test.go) inside_block:withInitScripts
<!--/codeinclude-->

<!--codeinclude-->
[Init script content](../../modules/clickhouse/testdata/init-db.sh)
<!--/codeinclude-->

#### Zookeeper

Clusterized ClickHouse requires to start Zookeeper and pass link to it via `config.xml`.

<!--codeinclude-->
[Include zookeeper](../../modules/clickhouse/clickhouse_test.go) inside_block:withZookeeper
<!--/codeinclude-->

!!!warning
    The `WithZookeeper` option will `panic` if it's not possible to create the Zookeeper config file.

#### Custom configuration

If you need to set a custom configuration, the module provides the `WithConfigFile` option to pass the path to a custom configuration file in XML format.

<!--codeinclude-->
[XML config file](../../modules/clickhouse/testdata/config.xml)
<!--/codeinclude-->

In the case you want to pass a YAML configuration file, you can use the `WithYamlConfigFile` option.

<!--codeinclude-->
[YAML config file](../../modules/clickhouse/testdata/config.yaml)
<!--/codeinclude-->

### Container Methods

The ClickHouse container exposes the following methods:

#### ConnectionHost

This method returns the host and port of the ClickHouse container, using the default, native `9000/tcp` port. E.g. `localhost:9000`

<!--codeinclude-->
[Get connection host](../../modules/clickhouse/clickhouse_test.go) inside_block:connectionHost
<!--/codeinclude-->

#### ConnectionString

This method returns the dsn connection string to connect to the ClickHouse container, using the default, native `9000/tcp` port obtained from the `ConnectionHost` method.
It's possible to pass extra parameters to the connection string, e.g. `dial_timeout=300ms` or `skip_verify=false`, in a variadic way.

e.g. `clickhouse://default:pass@localhost:9000?dial_timeout=300ms&skip_verify=false`

<!--codeinclude-->
[Get connection string](../../modules/clickhouse/clickhouse_test.go) inside_block:connectionString
<!--/codeinclude-->
