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
The **RunContainer** function is the main entry point to create a new VaultContainer instance. 
It takes a context and zero or more Option values to configure the container.
<!--codeinclude-->
[Creating a Vault container](../../modules/vault/vault_test.go) inside_block:RunContainer
<!--/codeinclude-->

### Use CLI to read data from Vault container:
<!--codeinclude-->
[Use CLI to read data](../../modules/vault/vault_test.go) inside_block:TestVaultGetSecretPathWithCLI
<!--/codeinclude-->

### Use HTTP API to read data from Vault container:
<!--codeinclude-->
[Use HTTP API to read data](../../modules/vault/vault_test.go) inside_block:TestVaultGetSecretPathWithHTTP
<!--/codeinclude-->

### Use client library to read data from Vault container:
Add Vault Client module to your Go dependencies:
```
go get -u github.com/hashicorp/vault-client-go
```
<!--codeinclude-->
[Use library to read data](../../modules/vault/vault_test.go) inside_block:TestVaultGetSecretPathWithClient
<!--/codeinclude-->

## Module Reference

The Vault module exposes one entrypoint function to create the containerr, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*VaultContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Vault container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Vault Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Vault. E.g. `testcontainers.WithImage("hashicorp/vault:1.13.0")`.

!!!info
    Default image name is `hashicorp/vault:1.13.0`.

<!--codeinclude-->
[Set image name](../../modules/vault/vault_test.go) inside_block:WithImageName
<!--/codeinclude-->

#### Wait Strategies

If you need to set a different wait strategy for Vault, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Vault.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Vault, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

#### Token

If you need to add token authentication, you can use the `WithToken`.
<!--codeinclude-->
[Add token authentication](../../modules/vault/vault_test.go) inside_block:WithToken
<!--/codeinclude-->

#### Command

If you need to run vault command in the container, you can use the `WithInitCommand`.
<!--codeinclude-->
[Run init command](../../modules/vault/vault_test.go) inside_block:WithInitCommand
<!--/codeinclude-->

### Container Methods

#### HttpHostAddress

This method returns the http host address of Vault, in the `http://<host>:<port>` format.

<!--codeinclude-->
[Get the HTTP host address](../../modules/vault/vault_test.go) inside_block:httpHostAddress
<!--/codeinclude-->
