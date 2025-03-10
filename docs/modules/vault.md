# Vault

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for Vault. [Vault](https://www.vaultproject.io/) is an open-source tool designed for securely storing, accessing, and managing secrets and sensitive data such as passwords, certificates, API keys, and other confidential information.

## Adding this module to your project dependencies

Please run the following command to add the Vault module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/vault
```

## Usage example
The **RunWithImage** function is the main entry point to create a new VaultContainer instance. 
It takes a context and zero or more Option values to configure the container.

<!--codeinclude-->
[Creating a Vault container](../../modules/vault/examples_test.go) inside_block:runVaultContainer
<!--/codeinclude-->

### Use CLI to read data from Vault container:
<!--codeinclude-->
[Use CLI to read data](../../modules/vault/vault_test.go) inside_block:containerCliRead
<!--/codeinclude-->

The `vaultContainer` is the container instance obtained from `RunWithImage`.

### Use HTTP API to read data from Vault container:
<!--codeinclude-->
[Use HTTP API to read data](../../modules/vault/vault_test.go) inside_block:httpRead
<!--/codeinclude-->

The `hostAddress` is obtained from the container instance. Please see [here](#httphostaddress) for more details.

### Use client library to read data from Vault container:
Add Vault Client module to your Go dependencies:

```
go get -u github.com/hashicorp/vault-client-go
```
<!--codeinclude-->
[Use library to read data](../../modules/vault/vault_test.go) inside_block:clientLibRead
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Vault module exposes one entrypoint function to create the container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*VaultContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Vault container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "hashicorp/vault:1.13.0")`.

{% include "../features/common_functional_options.md" %}

#### Token

If you need to add token authentication, you can use the `WithToken`.
<!--codeinclude-->
[Add token authentication](../../modules/vault/vault_test.go) inside_block:WithToken
<!--/codeinclude-->

#### Command

If you need to run a vault command in the container, you can use the `WithInitCommand`.
<!--codeinclude-->
[Run init command](../../modules/vault/vault_test.go) inside_block:WithInitCommand
<!--/codeinclude-->

### Container Methods

#### HttpHostAddress

This method returns the http host address of Vault, in the `http://<host>:<port>` format.

<!--codeinclude-->
[Get the HTTP host address](../../modules/vault/vault_test.go) inside_block:httpHostAddress
<!--/codeinclude-->
