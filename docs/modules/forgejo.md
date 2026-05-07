# Forgejo

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

## Introduction

The Testcontainers module for [Forgejo](https://forgejo.org/), a self-hosted Git forge. Forgejo is a community-driven fork of Gitea, providing a lightweight code hosting solution.

## Adding this module to your project dependencies

Please run the following command to add the Forgejo module to your Go dependencies:

```sh
go get github.com/testcontainers/testcontainers-go/modules/forgejo
```

## Usage example

<!--codeinclude-->
[Creating a Forgejo container](../../modules/forgejo/examples_test.go) inside_block:runForgejoContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

The Forgejo module exposes one entrypoint function to create the Forgejo container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "codeberg.org/forgejo/forgejo:11")`.

### Container Options

When starting the Forgejo container, you can pass options in a variadic way to configure it.

#### Admin Credentials

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Use `WithAdminCredentials(username, password, email)` to set the admin user credentials. An admin user is automatically created when the container starts. Default credentials are `forgejo-admin` / `forgejo-admin`.

#### Configuration via Environment

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

Use `WithConfig(section, key, value)` to set Forgejo configuration values using the `FORGEJO__section__key` environment variable format. See the [Forgejo Configuration Cheat Sheet](https://forgejo.org/docs/latest/admin/config-cheat-sheet/) for available options.

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Forgejo container exposes the following methods:

#### ConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

The `ConnectionString` method returns the HTTP URL for the Forgejo instance (e.g. `http://localhost:12345`).

#### SSHConnectionString

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.41.0"><span class="tc-version">:material-tag: v0.41.0</span></a>

The `SSHConnectionString` method returns the SSH endpoint for Git operations.
