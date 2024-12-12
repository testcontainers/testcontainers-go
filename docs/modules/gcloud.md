# GCloud

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>

## Introduction

The Testcontainers module for GCloud.

## Adding this module to your project dependencies

Please run the following command to add the GCloud module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/gcloud
```

## Usage example

!!!info
    By default, the all the emulators use `gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators` as the default Docker image, except for the BigQuery emulator, which uses `ghcr.io/goccy/bigquery-emulator:0.6.1`, and Spanner, which uses `gcr.io/cloud-spanner-emulator/emulator:1.4.0`.

### BigQuery

<!--codeinclude-->
[Creating a BigQuery container](../../modules/gcloud/bigquery_test.go) inside_block:runBigQueryContainer
[Obtaining a BigQuery client](../../modules/gcloud/bigquery_test.go) inside_block:bigQueryClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the client example above.

#### Data YAML (Seed File)

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you would like to do additional initialization in the BigQuery container, add a `data.yaml` file represented by an `io.Reader` to the container request with the `WithDataYAML` function.
That file is copied after the container is created but before it's started. The startup command then used will look like `--project test --data-from-yaml /testcontainers-data.yaml`.

An example of a `data.yaml` file that seeds the BigQuery instance with datasets and tables is shown below:

<!--codeinclude-->
[Data Yaml content](../../modules/gcloud/testdata/data.yaml)
<!--/codeinclude-->

!!!warning
    This feature is only available for the `BigQuery` container, and if you pass multiple `WithDataYAML` options, an error is returned.

### BigTable

<!--codeinclude-->
[Creating a BigTable container](../../modules/gcloud/bigtable_test.go) inside_block:runBigTableContainer
[Obtaining a BigTable Admin client](../../modules/gcloud/bigtable_test.go) inside_block:bigTableAdminClient
[Obtaining a BigTable client](../../modules/gcloud/bigtable_test.go) inside_block:bigTableClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the Admin client example above.

### Datastore

<!--codeinclude-->
[Creating a Datastore container](../../modules/gcloud/datastore_test.go) inside_block:runDatastoreContainer
[Obtaining a Datastore client](../../modules/gcloud/datastore_test.go) inside_block:datastoreClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the client example above.

### Firestore

<!--codeinclude-->
[Creating a Firestore container](../../modules/gcloud/firestore_test.go) inside_block:runFirestoreContainer
[Obtaining a Firestore client](../../modules/gcloud/firestore_test.go) inside_block:firestoreClient
<!--/codeinclude-->

It's important to set the target string of the `grpc.NewClient` method using the container's URI, as shown in the client example above.

### Pubsub

<!--codeinclude-->
[Creating a Pubsub container](../../modules/gcloud/pubsub_test.go) inside_block:runPubsubContainer
[Obtaining a Pubsub client](../../modules/gcloud/pubsub_test.go) inside_block:pubsubClient
<!--/codeinclude-->

It's important to set the target string of the `grpc.NewClient` method using the container's URI, as shown in the client example above.

### Spanner

<!--codeinclude-->
[Creating a Spanner container](../../modules/gcloud/spanner_test.go) inside_block:runSpannerContainer
[Obtaining a Spanner Admin client](../../modules/gcloud/spanner_test.go) inside_block:spannerAdminClient
[Obtaining a Spanner Database Admin client](../../modules/gcloud/spanner_test.go) inside_block:spannerDBAdminClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the Admin client example above.

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunXXXContainer(ctx, opts...)` functions are deprecated and will be removed in the next major release of _Testcontainers for Go_.

The GCloud module exposes one entrypoint function to create the different GCloud emulators, and each function receives three parameters:

```golang
func RunBigQuery(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*BigQueryContainer, error)
func RunBigTable(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*BigTableContainer, error)
func RunDatastore(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DatastoreContainer, error)
func RunFirestore(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*FirestoreContainer, error)
func RunPubsub(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*PubsubContainer, error)
func RunSpanner(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SpannerContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting any of the GCloud containers, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different GCloud Docker image, you can set a valid Docker image as the second argument in the `RunXXX` function (`RunBigQuery, RunDatastore`, ...).
E.g. `RunXXX(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The GCloud container exposes the following methods:
