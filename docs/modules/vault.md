# Vault

<img src="https://cdn.worldvectorlogo.com/logos/vault-enterprise.svg" width="200" />

Testcontainers module for Vault. [Vault](https://www.vaultproject.io/) is an open-source tool designed for securely storing, accessing, and managing secrets and sensitive data such as passwords, certificates, API keys, and other confidential information.

## Adding this module to your project dependencies

Please run the following command to add the Vault module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/vault
```

## Usage example
The **StartContainer** function is the main entry point to create a new VaultContainer instance. 
It takes a context and zero or more Option values to configure the container.
<!--codeinclude-->
[Creating a Vault container](../../modules/vault/vault_test.go) inside_block:StartContainer
<!--/codeinclude-->

### Use CLI to read data from Vault container:
<!--codeinclude-->
[Use CLI to read data](../../modules/vault/vault_test.go) inside_block:TestVaultFirstSecretPathWithCLI
<!--/codeinclude-->

### Use HTTP API to read data from Vault container:
<!--codeinclude-->
[Use HTTP API to read data](../../modules/vault/vault_test.go) inside_block:TestVaultFirstSecretPathWithHTTP
<!--/codeinclude-->

### Use client library to read data from Vault container:
<!--codeinclude-->
[Use library to read data](../../modules/vault/vault_test.go) inside_block:TestVaultFirstSecretPathWithClient
<!--/codeinclude-->