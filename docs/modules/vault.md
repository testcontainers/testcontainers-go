# Vault

<img src="https://cdn.worldvectorlogo.com/logos/vault-enterprise.svg" width="200" />

Testcontainers module for Vault. [Vault](https://www.vaultproject.io/) is an open-source tool designed for securely storing, accessing, and managing secrets and sensitive data such as passwords, certificates, API keys, and other confidential information.

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

## Container Options

You can set below options to create Vault container.

### Image 
If you need to set a different Vault image, you can use the `testcontainers.WithImage`. 

!!!info
    Default image name is `hashicorp/vault:1.13.0`.

<!--codeinclude-->
[Set image name](../../modules/vault/vault_test.go) inside_block:WithImageName
<!--/codeinclude-->

### Token
If you need to add token authentication, you can use the `WithToken`.
<!--codeinclude-->
[Add token authentication](../../modules/vault/vault_test.go) inside_block:WithToken
<!--/codeinclude-->

### Command
If you need to run vault command in the container, you can use the `WithInitCommand`.
<!--codeinclude-->
[Run init command](../../modules/vault/vault_test.go) inside_block:WithInitCommand
<!--/codeinclude-->