# Dex

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for Dex.

## Adding this module to your project dependencies

Please run the following command to add the Dex module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dex
```

## Usage example

<!--codeinclude-->
[Creating a Dex container](../../modules/dex/examples_test.go) inside_block:ExampleRun
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Dex module exposes one entrypoint function to create the Dex container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "dexidp/dex:v2.43.1-distroless")`.

### Container Options

When starting the Dex container, you can pass options in a variadic way to configure it.

#### Dex OAuth2 options

You can configure the OAuth2 issuer by setting `dex.WithIssuer`.

#### Dex logging options

You can configure the Dex logging by setting `dex.WithLogLevel`.
Valid log levels can be found in the [`log/slog`](https://pkg.go.dev/log/slog#Level) package documentation.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Dex container exposes the following methods:

#### RawOpenIDConfiguration

Fetch the OpenID Connect discovery document from the Dex container.
It patches all endpoint URLs to match the container HTTP endpoint.

#### OpenIDConfiguration

Fetch and parse the OpenID Connect discovery document and parse it.
Because it is using [RawOpenIDConfiguration](#rawopenidconfiguration) under the hood, the parsed endpoint URLs in the discovery document will match the mapped port of the Dex container.

#### CreateClientApp

Dynamically register a client application in the Dex instance by using the gRPC API.
ClientID and secret will be generated if not provided.

#### CreatePassword

Dynamically register a new 'password' (see also [`staticPasswords`](https://dexidp.io/docs/connectors/local/)) in Dex.
Either a plaintext password or a previously computed hash can be provided.
Dex expects passwords to be bcrypt encoded.
To compute compatible password hashed the [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) library from the experimental Go packages can be used.

**Note:** When logging in to Dex, the `email` is the primary identifier, **not** the `username`.
