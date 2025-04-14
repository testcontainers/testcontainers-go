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

The Google Cloud module exposes the following Go packages:

- [BigQuery](#bigquery): `github.com/testcontainers/testcontainers-go/modules/gcloud/bigquery`.
- [BigTable](#bigtable): `github.com/testcontainers/testcontainers-go/modules/gcloud/bigtable`.
- [Datastore](#datastore): `github.com/testcontainers/testcontainers-go/modules/gcloud/datastore`.
- [Firestore](#firestore): `github.com/testcontainers/testcontainers-go/modules/gcloud/firestore`.
- [Pubsub](#pubsub): `github.com/testcontainers/testcontainers-go/modules/gcloud/pubsub`.
- [Spanner](#spanner): `github.com/testcontainers/testcontainers-go/modules/gcloud/spanner`.
!!!info
    By default, the all the emulators use `gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators` as the default Docker image, except for the BigQuery emulator, which uses `ghcr.io/goccy/bigquery-emulator:0.6.1`, and Spanner, which uses `gcr.io/cloud-spanner-emulator/emulator:1.4.0`.

## BigQuery

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The BigQuery module exposes one entrypoint function to create the BigQuery container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the BigQuery container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "ghcr.io/goccy/bigquery-emulator:0.6.1")`.

{% include "./gcloud-shared.md" %}

#### Data YAML (Seed File)

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

If you would like to do additional initialization in the BigQuery container, add a `data.yaml` file represented by an `io.Reader` to the container request with the `WithDataYAML` function.
That file is copied after the container is created but before it's started. The startup command then used will look like `--project test --data-from-yaml /testcontainers-data.yaml`.

An example of a `data.yaml` file that seeds the BigQuery instance with datasets and tables is shown below:

<!--codeinclude-->
[Data Yaml content](../../modules/gcloud/bigquery/testdata/data.yaml)
<!--/codeinclude-->

### Examples

<!--codeinclude-->
[Creating a BigQuery container](../../modules/gcloud/bigquery/examples_test.go) inside_block:runBigQueryContainer
[Obtaining a BigQuery client](../../modules/gcloud/bigquery/examples_test.go) inside_block:bigQueryClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the client example above.

## BigTable

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The BigTable module exposes one entrypoint function to create the BigTable container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the BigTable container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "./gcloud-shared.md" %}

### Examples

<!--codeinclude-->
[Creating a BigTable container](../../modules/gcloud/bigtable/examples_test.go) inside_block:runBigTableContainer
[Obtaining a BigTable Admin client](../../modules/gcloud/bigtable/examples_test.go) inside_block:bigTableAdminClient
[Obtaining a BigTable client](../../modules/gcloud/bigtable/examples_test.go) inside_block:bigTableClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the Admin client example above.

## Datastore

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Datastore module exposes one entrypoint function to create the Datastore container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Datastore container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "./gcloud-shared.md" %}

### Examples

<!--codeinclude-->
[Creating a Datastore container](../../modules/gcloud/datastore/examples_test.go) inside_block:runDatastoreContainer
[Obtaining a Datastore client](../../modules/gcloud/datastore/examples_test.go) inside_block:datastoreClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the client example above.

## Firestore

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Firestore module exposes one entrypoint function to create the Firestore container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Firestore container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "./gcloud-shared.md" %}

### Examples

<!--codeinclude-->
[Creating a Firestore container](../../modules/gcloud/firestore/examples_test.go) inside_block:runFirestoreContainer
[Obtaining a Firestore client](../../modules/gcloud/firestore/examples_test.go) inside_block:firestoreClient
<!--/codeinclude-->

It's important to set the target string of the `grpc.NewClient` method using the container's URI, as shown in the client example above.

## Pubsub

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Pubsub module exposes one entrypoint function to create the Pubsub container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Pubsub container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "./gcloud-shared.md" %}

### Examples

<!--codeinclude-->
[Creating a Pubsub container](../../modules/gcloud/pubsub/examples_test.go) inside_block:runPubsubContainer
[Obtaining a Pubsub client](../../modules/gcloud/pubsub/examples_test.go) inside_block:pubsubClient
<!--/codeinclude-->

It's important to set the target string of the `grpc.NewClient` method using the container's URI, as shown in the client example above.

## Spanner

### Run function

- Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Spanner module exposes one entrypoint function to create the Spanner container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

### Container Options

When starting the Spanner container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators")`.

{% include "./gcloud-shared.md" %}

### Examples

<!--codeinclude-->
[Creating a Spanner container](../../modules/gcloud/spanner/examples_test.go) inside_block:runSpannerContainer
[Obtaining a Spanner Admin client](../../modules/gcloud/spanner/examples_test.go) inside_block:spannerAdminClient
[Obtaining a Spanner Database Admin client](../../modules/gcloud/spanner/examples_test.go) inside_block:spannerDBAdminClient
<!--/codeinclude-->

It's important to set the `option.WithEndpoint()` option using the container's URI, as shown in the Admin client example above.
