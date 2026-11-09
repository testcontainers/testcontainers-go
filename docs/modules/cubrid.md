# CUBRID

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [CUBRID](https://www.cubrid.org), an open source relational database.
It exposes a `cubrid://` connection string on the broker port `33000`.

## Adding this module to your project dependencies

Please run the following command to add the CUBRID module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/cubrid
```

## Usage example

<!--codeinclude-->
[Creating a CUBRID container](../../modules/cubrid/examples_test.go) inside_block:runCUBRIDContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The CUBRID module exposes one entrypoint function to create the CUBRID container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

`Run` returns once the broker and the database server have started, so only the image's default `CUBRID_COMPONENTS=ALL` is supported. If CUBRID cannot start, for example because of an unsupported locale, `Run` fails at once with the error lines from the container output instead of waiting for the startup timeout.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "cubrid/cubrid:11.4")`.

The `cubrid/cubrid` images are published for `linux/amd64` only, so on Apple Silicon they run under emulation and start more slowly.

### Container Options

When starting the CUBRID container, you can pass options in a variadic way to configure it.

#### WithDatabase

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `CUBRID_DB` environment variable, the name of the database created on startup. The default is `testdb`. An empty name is rejected.

```golang
cubrid.WithDatabase("mydb")
```

#### WithUsername

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `CUBRID_USER` environment variable, the user created on startup. The default is `testuser`. An empty name is rejected, and so are `dba` and `public`, which already exist in CUBRID.

```golang
cubrid.WithUsername("myuser")
```

#### WithPassword

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `CUBRID_PASSWORD` environment variable, the password of that user. The default is `testpass`. It cannot contain a single quote (`'`): the image puts the password in a SQL statement without escaping it, so the container fails to start.

```golang
cubrid.WithPassword("mypassword")
```

#### WithLocale

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Sets the `CUBRID_LOCALE` environment variable, the locale the database is created with, as `<language>_<territory>[.<charset>]`. The locale cannot be changed once the database exists. An empty value is rejected.

Without this option the image creates an `en_US` database, whose charset is ISO-8859-1. Korean and other text outside ISO-8859-1 then comes back garbled without an error, so use a UTF-8 locale for such text.

```golang
cubrid.WithLocale("ko_KR.utf8")
```

The image has these locales built in: `en_US.utf8`, `en_US.iso88591`, `ko_KR.utf8`, `ko_KR.euckr`, `ko_KR.iso88591`, `tr_TR.utf8` and `tr_TR.iso88591`. Without a charset, `en_US` means `en_US.iso88591`, while `ko_KR` and `tr_TR` mean `ko_KR.utf8` and `tr_TR.utf8`. Other locales make the container fail to start.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionString

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

This method returns the URL of the database, `cubrid://user:password@host:port/database`, which is the form used by pure Go CUBRID drivers. Extra arguments are appended as query parameters, such as `autocommit=false`.

<!--codeinclude-->
[Get connection string](../../modules/cubrid/cubrid_test.go) inside_block:connectionString
<!--/codeinclude-->

Drivers built on CUBRID's C interface (CCI) take another format, `cci:CUBRID:host:port:database:user:password:`. Build it from the container's `Host` and `MappedPort(ctx, "33000/tcp")`.

### Stopping the container

The `cubrid/cubrid` image ignores `SIGTERM`, so stopping the container takes the whole stop timeout, 10 seconds by default. If your tests do not need a graceful shutdown, use `testcontainers.TerminateContainer(ctr, testcontainers.StopTimeout(0))` or `testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))`.
