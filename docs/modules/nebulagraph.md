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

### RunCluster function

The NebulaGraph module provides a function to create a complete NebulaGraph cluster within a Docker network:

```golang
func RunCluster(ctx context.Context,
    graphdImg string, graphdCustomizers []testcontainers.ContainerCustomizer,
    storagedImg string, storagedCustomizers []testcontainers.ContainerCustomizer,
    metadImg string, metadCustomizers []testcontainers.ContainerCustomizer,
) (*NebulaGraphCluster, error)
```

This function creates a complete NebulaGraph cluster with customizable settings. It returns a `NebulaGraphCluster` struct that contains references to all four components:
- Meta Service (metad)
- Storage Service (storaged)
- Graph Service (graphd)
- Storage Activator (for registering storage with meta service)

### Container Options

The module supports customization for each service container (Meta, Storage, Graph, and Activator) through ContainerCustomizer options. Common customizations include:

- Custom images for each service
- Environment variables
- Resource limits
- Network settings
- Volume mounts
- Wait strategies

### Container Methods

The `NebulaGraphCluster` struct provides the following methods:

#### ConnectionString

```golang
func (c *NebulaGraphCluster) ConnectionString(ctx context.Context) (string, error)
```

Returns the host:port string for connecting to the NebulaGraph graph service (graphd).

#### Terminate

```golang
func (c *NebulaGraphCluster) Terminate(ctx context.Context) error
```

Stops and removes all containers in the NebulaGraph cluster (Meta, Storage, Graph, and Activator services) and cleans up the associated Docker network.

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
- Meta Service: HTTP health check on `/status` endpoint (port 19559)
- Graph Service: HTTP health check on `/status` endpoint (port 19669)
- Storage Service: Log-based health check for initialization
- Activator Service: Log-based health check and exit status for storage registration

A cluster is considered ready when:
1. Meta service is healthy and accessible
2. Graph service is healthy and accessible
3. Storage service is initialized and running
4. Storage service is successfully registered with the meta service via the activator
