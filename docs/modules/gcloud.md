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
    By default, the all the emulators use `gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators` as the default Docker image, except for the BigQuery emulator, which uses `ghcr.io/goccy/bigquery-emulator:0.4.3`, and Spanner, which uses `gcr.io/cloud-spanner-emulator/emulator:1.4.0`.

### BigQuery

<!--codeinclude-->
[Creating a BigQuery container](../../modules/gcloud/bigquery_test.go) inside_block:runBigQueryContainer
[Obtaining a BigQuery client](../../modules/gcloud/bigquery_test.go) inside_block:bigQueryClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the client example above.

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

It's important to set the target string of the `grpc.Dial` method using the container's URI, as shown in the client example above.

### Pubsub

<!--codeinclude-->
[Creating a Pubsub container](../../modules/gcloud/pubsub_test.go) inside_block:runPubsubContainer
[Obtaining a Pubsub client](../../modules/gcloud/pubsub_test.go) inside_block:pubsubClient
<!--/codeinclude-->

It's important to set the target string of the `grpc.Dial` method using the container's URI, as shown in the client example above.

### Spanner

<!--codeinclude-->
[Creating a Spanner container](../../modules/gcloud/spanner_test.go) inside_block:runSpannerContainer
[Obtaining a Spanner Admin client](../../modules/gcloud/spanner_test.go) inside_block:spannerAdminClient
[Obtaining a Spanner Database Admin client](../../modules/gcloud/spanner_test.go) inside_block:spannerDBAdminClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the Admin client example above.

## Module reference

The GCloud module exposes one entrypoint function to create the different GCloud emulators, and each function receives two parameters:

```golang
func RunBigQueryContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*BigQueryContainer, error)
func RunBigTableContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*BigTableContainer, error)
func RunDatastoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DatastoreContainer, error)
func RunFirestoreContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*FirestoreContainer, error)
func RunPubsubContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*PubsubContainer, error)
func RunSpannerContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SpannerContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting any of the GCloud containers, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different GCloud Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for GCloud. E.g. `testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "../features/common_functional_options.md" %}

### Container Methods

The GCloud container exposes the following methods:
