# MS SQL Server

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

## Introduction

The Testcontainers module for MS SQL Server.

## Adding this module to your project dependencies

Please run the following command to add the MS SQL Server module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/mssql
```

## Usage example

<!--codeinclude-->
[Creating a MS SQL Server container](../../modules/mssql/examples_test.go) inside_block:runMSSQLServerContainer
<!--/codeinclude-->

!!! warning "EULA Acceptance"
    Due to licensing restrictions you are required to explicitly accept an End User License Agreement (EULA) for the MS SQL Server container image. This is facilitated through the `WithAcceptEULA` function.

    Please see the [`microsoft-mssql-server` image documentation](https://hub.docker.com/_/microsoft-mssql-server#environment-variables) for a link to the EULA document.

## Module reference

The MS SQL Server module exposes one entrypoint function to create the MS SQL Server container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the MS SQL Server container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different MS SQL Server Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for MS SQL Server. E.g. `testcontainers.WithImage("mcr.microsoft.com/mssql/server:2022-RTM-GDR1-ubuntu-20.04")`.

#### End User License Agreement

Due to licensing restrictions you are required to explicitly accept an EULA for this container image. To do so, you must use the function `mssql.WithAcceptEula()`. Failure to include this will result in the container failing to start.

#### Password

If you need to set a different MS SQL Server password, you can use `mssql.WithPassword` with a valid password for MS SQL Server. E.g. `mssql.WithPassword("SuperStrong@Passw0rd")`.

!!!info
    If you set a custom password string, it must adhere to the MS SQL Server [Password Policy](https://learn.microsoft.com/en-us/sql/relational-databases/security/password-policy?view=sql-server-ver16).

{% include "../features/common_functional_options.md" %}

### Container Methods

The MS SQL Server container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the Microsoft SQL Server container, using the default `1433` port.
It's possible to pass extra parameters to the connection string, e.g. `encrypt=false` or `TrustServerCertificate=true`, in a variadic way.

```golang
connectionString, err := container.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
```
