# NebulaGraph Module

Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [NebulaGraph](https://nebula-graph.io/), a distributed, scalable, and lightning-fast graph database. This module manages a complete NebulaGraph cluster including Meta Service, Storage Service, and Graph Service components.

## Adding this module to your project dependencies

Add the NebulaGraph module to your Go dependencies:

```go
go get github.com/testcontainers/testcontainers-go/modules/nebulagraph
```

## Usage example

<!--codeinclude-->
[Creating a NebulaGraph container](../../modules/nebulagraph/nebulagraph_test.go) inside_block:TestNebulaGraphContainer
<!--/codeinclude-->

## Module Reference

### RunContainer function

The NebulaGraph module provides a simple entrypoint function to create a complete NebulaGraph cluster:

```golang
func RunContainer(ctx context.Context) (*NebulaGraphContainer, error)
```

This function creates a complete NebulaGraph cluster with default settings. It returns a `NebulaGraphContainer` struct that contains references to all three services (Meta, Storage, and Graph).

### Run function

For more advanced configuration, use the `Run` function which accepts customization options for each service:

```golang
func Run(ctx context.Context, 
    graphdCustomizers []testcontainers.ContainerCustomizer,
    storagedCustomizers []testcontainers.ContainerCustomizer,
    metadCustomizers []testcontainers.ContainerCustomizer,
) (*NebulaGraphContainer, error)
```

### Container Options

The module supports customization for each service container (Meta, Storage, and Graph) through ContainerCustomizer options. Common customizations include:

- Custom images
- Environment variables
- Resource limits
- Network settings
- Volume mounts

### Container Methods

The `NebulaGraphContainer` struct provides the following methods:

#### ConnectionString

```golang
func (c *NebulaGraphContainer) ConnectionString(ctx context.Context) (string, error)
```

Returns the host:port string for connecting to the NebulaGraph graph service (graphd).

#### Terminate

```golang
func (c *NebulaGraphContainer) Terminate(ctx context.Context) error
```

Stops and removes all containers in the NebulaGraph cluster (Meta, Storage, and Graph services).

## Default Configuration

The module uses the following default configurations:

- Default Images:
  - Graph Service: `vesoft/nebula-graphd:v3.8.0`
  - Meta Service: `vesoft/nebula-metad:v3.8.0`
  - Storage Service: `vesoft/nebula-storaged:v3.8.0`
  
- Exposed Ports:
  - Graph Service: 9669 (TCP), 19669 (HTTP)
  - Meta Service: 9559 (TCP), 19559 (HTTP)
  - Storage Service: 9779 (TCP), 19779 (HTTP)

## Health Checks

The module implements health checks for all services:
- Meta Service and Graph Service: HTTP health check on their respective status endpoints
- Storage Service: Log-based health check and registration validation

A container is considered ready when all services are healthy and the storage service is properly registered with the meta service.
