# Docker MCP Gateway

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for the Docker MCP Gateway.

## Adding this module to your project dependencies

Please run the following command to add the Docker MCP Gateway module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/dockermcpgateway
```

## Usage example

<!--codeinclude-->
[Creating a DockerMCPGateway container](../../modules/dockermcpgateway/examples_test.go) inside_block:run_mcp_gateway
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The DockerMCPGateway module exposes one entrypoint function to create the DockerMCPGateway container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "docker/mcp-gateway:latest")`.

### Container Options

When starting the DockerMCPGateway container, you can pass options in a variadic way to configure it.

{% include "../features/common_functional_options_list.md" %}

#### WithTools

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithTools` option to set the tools from a server to be available in the MCP Gateway container. Adding multiple tools for the same server will append to the existing tools for that server, and no duplicate tools will be added for the same server.

```golang
dockermcpgateway.WithTools("brave", []string{"brave_local_search", "brave_web_search"})
```

#### WithSecrets

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Use the `WithSecrets` option to set the tools from a server to be available in the MCP Gateway container. Empty keys are not allowed, although empty values are allowed for a key.

```golang
dockermcpgateway.WithSecret("github_token", "test_value")
dockermcpgateway.WithSecrets(map[string]{
    "github_token": "test_value",
    "foo":          "bar",
})
```

### Container Methods

The DockerMCPGateway container exposes the following methods:

#### Tools

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns a map of tools available in the MCP Gateway container, where the key is the server name and the value is a slice of tool names.

```golang
tools := ctr.Tools()
```

#### GatewayEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the endpoint of the MCP Gateway container, which is a string containing the host and mapped port for the default MCP Gateway port (8811/tcp).

```golang
endpoint := ctr.GatewayEndpoint()
```
### Examples

#### Connecting to the MCP Gateway using an MCP client

This example shows the usage of the MCP Gateway module to connect with an [MCP client](https://github.com/modelcontextprotocol/go-sdk).

<!--codeinclude-->
[Run the MCP Gateway](../../modules/dockermcpgateway/examples_test.go) inside_block:run_mcp_gateway
[Get MCP Gateway's endpoint](../../modules/dockermcpgateway/examples_test.go) inside_block:get_gateway
[Connect with an MCP client](../../modules/dockermcpgateway/examples_test.go) inside_block:connect_mcp_client
[List tools](../../modules/dockermcpgateway/examples_test.go) inside_block:list_tools
<!--/codeinclude-->
