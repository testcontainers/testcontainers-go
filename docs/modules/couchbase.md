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
[Creating a Couchbase container](../../modules/couchbase/examples_test.go) inside_block:runCouchbaseContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Since testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0"><span class="tc-version">:material-tag: v0.32.0</span></a>

!!!info
    The `RunContainer(ctx, opts...)` function is deprecated and will be removed in the next major release of _Testcontainers for Go_.

The Couchbase module exposes one entrypoint function to create the Couchbase container, and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CouchbaseContainer, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
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
* Finally, it will wait for all nodes to be healthy. Depending on the enabled services, it will use a different wait strategy to check if the node is healthy:
	- It will wait for the `/pools/default` endpoint in the management port to return a 200 HTTP status code and the response body to contain the `healthy` key set to `true`.
	- If the `Query` service is enabled, it will wait for the `/admin/ping` endpoint in the query port to return a 200 HTTP status code.
	- If the `Analytics` service is enabled, it will wait for the `/admin/ping` endpoint in the analytics port to return a 200 HTTP status code.
	- If the `Eventing` service is enabled, it will wait for the `/api/v1/config` endpoint in the eventing port to return a 200 HTTP status code.

### Container Ports

Here you can find the list with the default ports used by the Couchbase container. The Management ports (`MGMT_PORT` and `MGMT_SSL_PORT`) and the Service ports for `kv`, `query` and `search` are exposed by default.

!!!tip
	You can export the service ports for Analytics and Eventing by using the `WithServiceAnalytics` and `WithServiceEventing` optional functions.

<!--codeinclude-->
[Container Ports](../../modules/couchbase/couchbase.go) inside_block:containerPorts
<!--/codeinclude-->

### Container Options

When starting the Couchbase container, you can pass options in a variadic way to configure it.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "couchbase:6.5.1")`.

You can find the Docker images that are currently tested in this module, for the Enterprise and Community editions, in the following list:

<!--codeinclude-->
[Docker images](../../modules/couchbase/couchbase_test.go) inside_block:dockerImages
<!--/codeinclude-->

{% include "../features/common_functional_options.md" %}

#### Credentials

If you need to change the default credentials for the admin user, you can use `WithAdminCredentials(user, password)` with a valid username and password.
When the password has less than 6 characters, the container won't be created and the `New` function will throw an error.

!!!info
	In the case this optional function is not called, the default username is `Administrator` and the default password is `password`.

#### Buckets

When creating a new Couchbase container, you can create one or more buckets. The module exposes a `WithBuckets` optional function that accepts a slice of buckets to be created.
To create a new bucket, the module also exposes a `NewBucket` function, where you can pass the bucket name.

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

### Container Methods

#### ConnectionString

The `ConnectionString` method returns the connection string to connect to the Couchbase container instance. 
It returns a string with the format `couchbase://<host>:<port>`.

#### Username

The `Username` method returns the username of the Couchbase administrator. 

#### Password

The `Password` method returns the password of the Couchbase administrator.
