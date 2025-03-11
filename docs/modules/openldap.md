# OpenLDAP

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

## Introduction

The Testcontainers module for OpenLDAP.

## Adding this module to your project dependencies

Please run the following command to add the OpenLDAP module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/openldap
```

## Usage example

<!--codeinclude-->
[Creating a OpenLDAP container](../../modules/openldap/examples_test.go) inside_block:runOpenLDAPContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The OpenLDAP module exposes one entrypoint function to create the OpenLDAP container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenLDAP container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "bitnami/openldap:2.6.6")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The OpenLDAP container exposes the following methods:

#### ConnectionString

This method returns the connection string to connect to the OpenLDAP container, using the `1389` port.

<!--codeinclude-->
[Get connection string](../../modules/openldap/openldap_test.go) inside_block:connectionString
<!--/codeinclude-->

#### LoadLdif

This method loads an ldif file in the OpenLDAP server.
It returns and error if there is any problem with the ldif file loading process.

<!--codeinclude-->
[Load ldif](../../modules/openldap/openldap_test.go) inside_block:loadLdif
<!--/codeinclude-->

#### Initial Ldif

If you would like to load an ldif at the initialization of the openldap container, you can use `WithInitialLdif` function.
The file will be copied after the container is started and loaded in openldap.
