# CockroachDB

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for CockroachDB.

## Adding this module to your project dependencies

Please run the following command to add the CockroachDB module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/cockroachdb
```

## Usage example

<!--codeinclude-->
[Creating a CockroachDB container](../../modules/cockroachdb/examples_test.go) inside_block:runCockroachDBContainer
<!--/codeinclude-->

## Module reference

The CockroachDB module exposes one entrypoint function to create the CockroachDB container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

!!!warning
    When TLS is enabled there's a very small, unlikely chance that the underlying driver can panic when registering the driver as part of waiting for CockroachDB to be ready to accept connections. If this is repeatedly happening please open an issue.

### Container Options

When starting the CockroachDB container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different CockroachDB Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for CockroachDB. E.g. `testcontainers.WithImage("cockroachdb/cockroach:latest-v23.1")`.

{% include "../features/common_functional_options.md" %}

#### Database

Set the database that is created & dialled with `cockroachdb.WithDatabase`.

#### Password authentication

Disable insecure mode and connect with password authentication by setting `cockroachdb.WithUser` and `cockroachdb.WithPassword`.

#### Store size

Control the maximum amount of memory used for storage, by default this is 100% but can be changed by provided a valid option to `WithStoreSize`. Checkout https://www.cockroachlabs.com/docs/stable/cockroach-start#store for the full range of options available.

#### TLS authentication

`cockroachdb.WithTLS` lets you provide the CA certificate along with the certicate and key for the node & clients to connect with.
Internally CockroachDB requires a client certificate for the user to connect with.

A helper `cockroachdb.NewTLSConfig` exists to generate all of this for you.

### Container Methods

The CockroachDB container exposes the following methods:

#### ConnectionString

Dial address to open a new connection.

#### MustConnectionString

Same as `ConnectionString` but any error to generate the address will raise a panic

#### TLSConfig

Returns `*tls.Config` setup to allow you to dial your client over TLS, if enabled, else this will error with `cockroachdb.ErrTLSNotEnabled`.
