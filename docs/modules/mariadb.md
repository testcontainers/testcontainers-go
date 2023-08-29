# MariaDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for MariaDB.

## Adding this module to your project dependencies

Please run the following command to add the MariaDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mariadb
```

## Usage example

<!--codeinclude-->
[Creating a MariaDB container](../../modules/mariadb/mariadb.go)
<!--/codeinclude-->

<!--codeinclude-->
[Test for a MariaDB container](../../modules/mariadb/mariadb_test.go)
<!--/codeinclude-->

## Module reference

The MariaDB module exposes one entrypoint function to create the MariaDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MariaDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different MariaDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MariaDB. E.g. `testcontainers.WithImage("mariadb:10.5.5")`.

#### Wait Strategies

If you need to set a different wait strategy for MariaDB, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for MariaDB.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for MariaDB, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

### Container Methods

The MariaDB container exposes the following methods:
