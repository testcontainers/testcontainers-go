# InfluxDB

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

## Introduction

A testcontainers module for InfluxDB V1 and V2.

## Adding this module to your project dependencies

Please run the following command to add the InfluxDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/influxdb
```

## Usage example

### InfluxDB

<!--codeinclude--> 
[Creating an InfluxDB V1 container](../../modules/influxdb/examples_test.go) inside_block:runInfluxContainer
[Creating an InfluxDB V2 container](../../modules/influxdb/examples_test.go) inside_block:runInfluxV2Container
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The InfluxDB module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*InfluxDbContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "influxdb:1.8.0")`.

!!!info
    Note that `influxdb:latest` will pull a version 2 image.

### Container Options

When starting the container, you can pass options in a variadic way to configure it.

!!!tip

    You can find configuration information for the InfluxDB image on [Docker Hub](https://hub.docker.com/_/influxdb) and a list of possible 
    environment variables on [InfluxDB documentation](https://docs.influxdata.com/influxdb/v1/administration/config/).

#### Set username, password and database name

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

By default, authentication is disabled and no credentials are needed to use the Influx API against the test container.
If you want to test with credentials, include the appropriate environment variables to do so.

#### Configuring InfluxDB V2

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

When running the InfluxDB V2 image, you can override the default configuration by using options prefixed by `influxdb.WithV2`.
The following options are available:

- `WithV2(org, bucket string)`: Configures organization and bucket name. This option is required to run the InfluxDB V2 image.
- `WithV2Auth(org, bucket, username, password string)`: Sets the username and password for the initial user.
- `WithV2SecretsAuth(org, bucket, usernameFile, passwordFile string)`: Sets the username and password file path.
- `WithV2Retention(retention time.Duration)`: Sets the default bucket retention policy.
- `WithV2AdminToken(token string)`:  Sets the admin token for the initial user.
- `WithV2SecretsAdminToken(tokenFile string)`: Sets the admin token file path.

#### WithInitDb

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

While the InfluxDB image will obey the `/docker-entrypoint-initdb.d` directory as is common, that directory does not
exist in the default image. Instead, you can use the `WithInitDb` option to pass a directory which will be copied to
when the container starts. Any `*.sh` or `*.iql` files in the directory will be processed by the image upon startup.
When executing these scripts, the `init-influxdb.sh` script in the image will start the InfluxDB server, run the
scripts, stop the server, and restart the server.  This makes it tricky to detect the readiness of the container.
This module looks for that and adds some extra tests for readiness, but these could be fragile.

!!!important
    The `WithInitDb` option receives a path to the parent directory of one named `docker-entrypoint-initdb.d`. This is
    because the `docker-entrypoint-initdb.d` directory is not present in the image.

#### WithConfigFile

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

If you need to set a custom configuration, you can use `WithConfigFile` option to pass the path to a custom configuration file.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

#### ConnectionUrl

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

This function is a simple helper to return a URL to the container, using the default `8086` port.

<!--codeinclude-->
[ConnectionUrl](../../modules/influxdb/influxdb_test.go) inside_block:influxConnectionUrl
<!--/codeinclude-->

Please check the existence of two methods: `ConnectionUrl` and `MustConnectionUrl`. The latter is used to avoid the need to handle errors,
while the former is used to return the URL and the error. `MustConnectionUrl` will panic if an error occurs.

!!!info
    The `ConnectionUrl` and `MustConnectionUrl` methods only support HTTP connections at the moment.
