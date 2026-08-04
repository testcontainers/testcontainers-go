# Firebird

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Firebird, a relational database that runs on Linux, Windows, and Unix platforms.
It exposes a `firebird://` connection string on port `3050`.

## Adding this module to your project dependencies

Please run the following command to add the Firebird module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/firebird
```

## Usage example

<!--codeinclude-->
[Creating a Firebird container](../../modules/firebird/examples_test.go) inside_block:runFirebirdContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Firebird module exposes one entrypoint function to create the Firebird container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ghcr.io/jacobalberty/firebird:v3.0")`.

### Container Options

When starting the Firebird container, you can pass options in a variadic way to configure it.

!!!tip
    You can find all the available configuration and environment variables for the Firebird Docker image on [GitHub](https://github.com/jacobalberty/firebird-docker).

#### WithDatabase

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `FIREBIRD_DATABASE` environment variable. The default database name is `test.fdb`.

```golang
firebird.WithDatabase("mydb.fdb")
```

#### WithUsername

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `FIREBIRD_USER` environment variable. The default username is `test`.

```golang
firebird.WithUsername("myuser")
```

#### WithPassword

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `FIREBIRD_PASSWORD` environment variable. The default password is `test`.

```golang
firebird.WithPassword("mypassword")
```

#### WithSYSDBAPassword

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `ISC_PASSWORD` environment variable, which is the SYSDBA master password. The default is `masterkey`.

```golang
firebird.WithSYSDBAPassword("mysysdbapassword")
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the connection string to connect to the Firebird container, using the `firebird://` scheme on the default port `3050`.

<!--codeinclude-->
[Get connection string](../../modules/firebird/firebird_test.go) inside_block:connectionString
<!--/codeinclude-->
