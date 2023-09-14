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

## Module reference

The ClickHouse module exposes one entrypoint function to create the ClickHouse container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*ClickHouseContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Ports

Here you can find the list with the default ports used by the ClickHouse container.

<!--codeinclude-->
[Container Ports](../../modules/clickhouse/clickhouse.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the ClickHouse container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different ClickHouse Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for ClickHouse. E.g. `testcontainers.WithImage("clickhouse/clickhouse-server:23.3.8.21-alpine")`.

#### Wait Strategies

If you need to set a different wait strategy for ClickHouse, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for ClickHouse.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for ClickHouse, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

#### Startup Commands

!!!info
    Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Testcontainers exposes the `WithStartupCommand(e ...Executable)` option to run arbitrary commands in the container right after it's started.

!!!info
    To better understand how this feature works, please read the [Create containers: Lifecycle Hooks](../../features/creating_container/#lifecycle-hooks) documentation.

It also exports an `Executable` interface, defining one single method: `AsCommand()`, which returns a slice of strings to represent the command and positional arguments to be executed in the container.

You could use this feature to run a custom script, or to run a command that is not supported by the module right after the ClickHouse container is started.

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
