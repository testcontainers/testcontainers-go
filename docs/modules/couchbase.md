# Couchbase

Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>

## Introduction

The Testcontainers module for Couchbase.

## Adding this module to your project dependencies

Please run the following command to add the Couchbase module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/couchbase
```

## Usage example

<!--codeinclude-->
[Start Couchbase](../../modules/couchbase/couchbase_test.go) inside_block:withBucket
<!--/codeinclude-->

## Module Reference

The Couchbase module exposes one entrypoint function to create the Couchbase container, and this function receives two parameters:

```golang
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CouchbaseContainer, error)
```

- `context.Context`, the Go context.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing options.

Once the container is started, it will perform the following operations, **in this particular order**:

* Wait until the node is online, waiting for the `/pools` endpoint in the management port to return a 200 HTTP status code.
* Check for Enterprise services, sending a GET request to the `/pools` endpoint in the management port. If the response contains the `isEnterprise` key set to `false`, it will check if the Analytics or the Eventing services are enabled. If so, it will raise an error.
* Rename the node, sending a POST request to the `/node/controller/rename` endpoint in the management port.
* Initialize the services, sending a POST request to the `/node/controller/setupServices` endpoint in the management port, passing as body of the request the list of enabled services.
* Set the memory quotas, sending a POST request to the `/pools/default` endpoint in the management port, passing as body of the request the memory quota for each enabled service.
* Configure the Admin user, sending a POST request to the `/settings/web` endpoint in the management port, passing as body of the request the username and password of the admin user.
* Configure the external ports, sending a POST request to the `/node/controller/setupAlternateAddresses/external` endpoint in the management port, passing as body of the request the external mapped ports for each enabled service.
* If the `Index` service is enabled, configure the indexer, sending a POST request to the `/settings/indexes` endpoint in the management port, passing as body of the request the defined storage mode. If the Community Edition is used, it will make sure the storage mode is `forestdb`. If the Enterprise Edition is used, it will make sure the storage mode is not `forestdb`, changing to `memory_optimized` in that case.
* Finally, it will wait for all nodes to be healthy. Depending of the enabled services, it will use a different wait strategy to check if the node is healthy:
	- It will wait for the `/pools/default` endpoint in the management port to return a 200 HTTP status code and the response body to contain the `healthy` key set to `true`.
	- If the `Query` service is enabled, it will wait for the `/admin/ping` endpoint in the query port to return a 200 HTTP status code.
	- If the `Analytics` service is enabled, it will wait for the `/admin/ping` endpoint in the analytics port to return a 200 HTTP status code.
	- If the `Eventing` service is enabled, it will wait for the `/api/v1/config` endpoint in the eventing port to return a 200 HTTP status code.

### Container Ports

<!--codeinclude-->
[Container Ports](../../modules/couchbase/couchbase.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the Couchbase container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Couchbase Docker image, you can use `testcontainers.WithImage` with a valid Docker image
for Couchbase. E.g. `testcontainers.WithImage("docker.io/couchbase:6.5.1")`.

By default, the container will use the following Docker image:

<!--codeinclude-->
[Default Docker image](../../modules/couchbase/couchbase.go) inside_block:defaultImage
<!--/codeinclude-->

#### Wait Strategies

If you need to set a different wait strategy for Couchbase, you can use `testcontainers.WithWaitStrategy` with a valid wait strategy
for Couchbase.

!!!info
    The default deadline for the wait strategy is 60 seconds.

At the same time, it's possible to set a wait strategy and a custom deadline with `testcontainers.WithWaitStrategyAndDeadline`.

#### Docker type modifiers

If you need an advanced configuration for Couchbase, you can leverage the following Docker type modifiers:

- `testcontainers.WithConfigModifier`
- `testcontainers.WithHostConfigModifier`
- `testcontainers.WithEndpointSettingsModifier`

Please read the [Create containers: Advanced Settings](../features/creating_container.md#advanced-settings) documentation for more information.

#### Credentials

If you need to change the default credentials for the admin user, you can use `WithAdminCredentials(user, password)` with a valid username and password.
When the password has less than 6 characters, the container won't be created and the `RunContainer` function will throw an error.

!!!info
	The default username is `Administrator` and the default password is `password`.

#### Buckets

When creating a new Couchbase container, you can create one or more buckets. The module provides with a `WithBuckets` function that accepts an array of buckets to be created.
To create a new bucket, the module exposes a `NewBucket` function, where you can pass the bucket name.

It's possible to customize a newly created bucket, using the following options:

- `WithQuota`: sets the bucket quota in megabytes. The minimum value is 100 MB.
- `WithReplicas`: sets the number of replicas for this bucket. The minimum value is 0 and the maximum value is 3.
- `WithFlushEnabled`: sets whether the bucket should be flushed when the container is stopped.
- `WithPrimaryIndex`: sets whether the primary index should be created for this bucket.

<!--codeinclude-->
[Adding a new bucket](../../modules/couchbase/couchbase_test.go) inside_block:withBucket
<!--/codeinclude-->

#### Index Storage

It's possible to set the storage mode to be used for all global secondary indexes in the cluster.

!!!warning
	Please note: `plasma` and `memory optimized` are options in the **Enterprise Edition** of Couchbase Server. If you are using the Community Edition, the only value allowed is `forestdb`.

<!--codeinclude-->
[Storage types](../../modules/couchbase/storage_mode.go) inside_block:storageTypes
<!--/codeinclude-->

#### Services

By default, the container will start with the following services: `kv`, `n1ql`, `fts` and `index`.

!!!warning
	When running the Enterprise Edition of Couchbase Server, the module provides two functions to enable or disable services:
	`WithServiceAnalytics` and `WithServiceEventing`. Else, it will throw an error and the container won't be created.

<!--codeinclude-->
[Docker images](../../modules/couchbase/couchbase_test.go) inside_block:dockerImages
<!--/codeinclude-->

### Container Methods

#### ConnectionString

The `ConnectionString` method returns the connection string to connect to the Couchbase container instance. 
It returns a string with the format `couchbase://<host>:<port>`.

<!--codeinclude-->
[Connect to Couchbase](../../modules/couchbase/couchbase_test.go) inside_block:connectToCluster
<!--/codeinclude-->

#### Username

The `Username` method returns the username of the Couchbase administrator. 

<!--codeinclude-->
[Connect to Couchbase using Credentials](../../modules/couchbase/couchbase_test.go) inside_block:getCredentials
<!--/codeinclude-->

#### Password

The `Password` method returns the password of the Couchbase administrator.

<!--codeinclude-->
[Connect to Couchbase using Credentials](../../modules/couchbase/couchbase_test.go) inside_block:getCredentials
<!--/codeinclude-->
