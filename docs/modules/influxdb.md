# InfluxDB

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.30.0"><span class="tc-version">:material-tag: v0.30.0</span></a>

## Introduction

A testcontainers module for InfluxDB.  This module supports v1.x of InfluxDB.   

## Adding this module to your project dependencies

Please run the following command to add the InfluxDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/influxdb
```

## Usage example

<!--codeinclude--> 
[Creating an InfluxDB container](../../modules/influxdb/examples_test.go) inside_block:runInfluxContainer
<!--/codeinclude-->

## Module Reference

The InfluxDB module exposes one entrypoint function to create the container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InfluxDbContainer, error) {}
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the container, you can pass options in a variadic way to configure it.

!!!tip

    You can find configuration information for the InfluxDB image on [Docker Hub](https://hub.docker.com/_/influxdb) and a list of possible 
    environment variables on [InfluxDB documentation](https://docs.influxdata.com/influxdb/v1/administration/config/).

#### Image

To use a different Docker image, you can use the `testcontainers.WithImage` option to specify the 
image, E.g. `testcontainers.WithImage("influxdb:1.8.0")`.  By default, the 1.8.10 image is used.  Note that
`influxdb:latest` will get you a version 2 image which is not supported by this module.


{% include "../features/common_functional_options.md" %}

#### Set username, password and database name

By default, authentication is disabled and no credentials are needed to use the Influx API against the test container.
If you want to test with credentials, include the appropriate environment variables to do so.

#### Init Scripts

While the InfluxDB image will obey the `/docker-entrypoint-initdb.d` directory as is common, that directory does not
exist in the default image. Instead, you can use the `WithInitDb` option to pass a directory which will be copied to
when the container starts. Any `*.sh` or `*.iql` files in the directory will be processed by the image upon startup.
When executing these scripts, the `init-influxdb.sh` script in the image will start the InfluxDB server, run the
scripts, stop the server, and restart the server.  This makes it tricky to detect the readiness of the container.
This module looks for that and adds some extra tests for readiness, but these could be fragile.

!!!important
    The `WithInitDb` option receives a path to the parent directory of one named `docker-entrypoint-initdb.d`. This is
    because the `docker-entrypoint-initdb.d` directory is not present in the image.

#### Custom configuration

If you need to set a custom configuration, you can use `WithConfigFile` option to pass the path to a custom configuration file.

### Container Methods

#### ConnectionUrl

This function is a simple helper to return a URL to the container, using the default `8086` port.

<!--codeinclude-->
[ConnectionUrl](../../modules/influxdb/influxdb_test.go) inside_block:influxConnectionUrl
<!--/codeinclude-->

Please check the existence of two methods: `ConnectionUrl` and `MustConnectionUrl`. The latter is used to avoid the need to handle errors,
while the former is used to return the URL and the error. `MustConnectionUrl` will panic if an error occurs.

!!!info
    The `ConnectionUrl` and `MustConnectionUrl` methods only support HTTP connections at the moment.
