# Neo4j

The Testcontainers module for [Neo4j](https://neo4j.com/), the leading graph platform.

## Adding this module to your project dependencies

Please run the following command to add the Neo4j module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/neo4j
```

## Usage example

Running Neo4j as a single-instance server, with the [APOC plugin](https://neo4j.com/developer/neo4j-apoc/) enabled:

<!--codeinclude-->
[Creating a Neo4j container](../../modules/neo4j/neo4j_test.go) inside_block:neo4jCreateContainer
<!--/codeinclude-->