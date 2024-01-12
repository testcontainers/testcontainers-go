# OpenLDAP

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

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

## Module reference

The OpenLDAP module exposes one entrypoint function to create the OpenLDAP container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenLDAPContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the OpenLDAP container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different OpenLDAP Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for OpenLDAP. E.g. `testcontainers.WithImage("bitnami/openldap:2.6.6")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The OpenLDAP container exposes the following methods:
