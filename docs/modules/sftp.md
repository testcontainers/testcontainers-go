# SFTP

Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

## Introduction

The Testcontainers module for [SFTP](https://github.com/atmoz/sftp) provides an easy-to-use SFTP server built on OpenSSH. It is ideal for testing file-transfer workflows without connecting to a real remote server.

## Adding this module to your project dependencies

Please run the following command to add the SFTP module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/sftp
```

## Usage example

<!--codeinclude-->
[Creating a SFTP container](../../modules/sftp/examples_test.go) inside_block:runSFTPContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

!!!info
    The `Run` function requires at least one user to be configured via `WithUser`, otherwise it returns an error.

The SFTP module exposes one entrypoint function to create the SFTP container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "atmoz/sftp:latest", sftp.WithUser("alice", "secret"))`.

### Container Options

When starting the SFTP container, you can pass options in a variadic way to configure it.

#### WithUser

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

`WithUser(username, password string)` adds an SFTP user. At least one user is required. Multiple calls accumulate users.

```golang
sftpContainer, err := sftp.Run(ctx, "atmoz/sftp:latest",
    sftp.WithUser("alice", "secret"),
    sftp.WithUser("bob", "password"),
)
```

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The SFTP container exposes the following methods:

#### Address

- Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.44.0"><span class="tc-version">:material-tag: v0.44.0</span></a>

`Address(ctx context.Context) (string, error)` returns the `host:port` address of the SFTP server. Use this value directly with an SSH or SFTP client dial call.

```golang
addr, err := sftpContainer.Address(ctx)
// addr is e.g. "localhost:49160"
```
