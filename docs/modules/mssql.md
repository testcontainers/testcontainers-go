# MSSQLServer

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for MSSQLServer.

## Adding this module to your project dependencies

Please run the following command to add the MSSQLServer module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mssql
```

## Usage example

<!--codeinclude-->

[Creating a MSSQLServer container](../../modules/mssql/examples_test.go) inside_block:runMSSQLServerContainer

<!--/codeinclude-->

!!! warning "EULA Acceptance"
Due to licencing restrictions you are required to accept an End User License Agreement (EULA) for this container image. To indicate that you accept the MS SQL Server image EULA, call the `WithAcceptEULA("Y")` method passing a "yes" parameter as `"Y"` to indicate EULA acceptance.

    Please see the [`microsoft-mssql-server` image documentation](https://hub.docker.com/_/microsoft-mssql-server#environment-variables) for a link to the EULA document.

## Module reference

The MSSQLServer module exposes one entrypoint function to create the MSSQLServer container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MSSQLServer container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different MSSQLServer Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MSSQLServer. E.g. `testcontainers.WithImage("mcr.microsoft.com/mssql/server:2022-latest")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The MSSQLServer container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Microsoft SQL Server container, using the default `1433` port.
It's possible to pass extra parameters to the connection string, e.g. `encrypt=false` or `TrustServerCertificate=true`, in a variadic way.

<!--codeinclude-->

[Get connection string](../../modules/mssql/mssql_test.go) inside_block:connectionString

<!--/codeinclude-->
