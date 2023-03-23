# Couchbase

<img src="https://cdn.worldvectorlogo.com/logos/couchbase.svg" width="300" />

Testcontainers module for Couchbase. [Couchbase](https://www.couchbase.com/) is a document oriented NoSQL database.

## Adding this module to your project dependencies

Please run the following command to add the Couchbase module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/couchbase
```

## Usage example

1. The **StartContainer** function is the main entry point to create a new CouchbaseContainer instance. 
It takes a context and zero or more Option values to configure the container. 
It creates a new container instance, initializes the couchbase cluster, and creates buckets. 
If successful, it returns the **CouchbaseContainer** instance.

<!--codeinclude-->
[Start Couchbase](../../modules/couchbase/couchbase_test.go) inside_block:withBucket
<!--/codeinclude-->

2. The **ConnectionString** method returns the connection string to connect to the Couchbase container instance. 
It returns a string with the format `couchbase://<host>:<port>`.
The **Username** method returns the username of the Couchbase administrator. 
The **Password** method returns the password of the Couchbase administrator.

<!--codeinclude-->
[Connect to Couchbase](../../modules/couchbase/couchbase_test.go) inside_block:connectToCluster
<!--/codeinclude-->

## Module Reference

The Couchbase module exposes one entrypoint function to create the Couchbase container, and this function receives two parameters:

```golang
func StartContainer(ctx context.Context, opts ...Option) (*CouchbaseContainer, error)
```

- `context.Context`, the Go context.
- `Option`, a variad argument for passing options.

### Container Options

When starting the Couchbase container, you can pass options in a variadic way to configure it.

#### Image

If you need to set a different Couchbase Docker image, you can use `WithImageName` with a valid Docker image
for Couchbase. E.g. `WithImageName("docker.io/couchbase:6.5.1")`.

By default, the container will use the following Docker image:

<!--codeinclude-->
[Default Docker image](../../modules/couchbase/couchbase.go) inside_block:defaultImage
<!--/codeinclude-->

#### Credentials

If you need to change the default credentials for the admin user, you can use `WithCredentials(user, password)` with a valid username and password.

!!!info
	The default username is `Administrator` and the default password is `password`.

#### Bucket

When creating a new Couchbase container, you can create one or more buckets. The module provides a `NewBucket` function to create a new bucket, where
you can pass the bucket name.

<!--codeinclude-->
[Adding a new bucket](../../modules/couchbase/couchbase_test.go) inside_block:withBucket
<!--/codeinclude-->

It's possible to customize a newly created bucket, using the following options:

- `WithQuota`: sets the bucket quota in megabytes. The minimum value is 100 MB.
- `WithReplicas`: sets the number of replicas for this bucket. The minimum value is 0 and the maximum value is 3.
- `WithFlushEnabled`: sets whether the bucket should be flushed when the container is stopped.
- `WithPrimaryIndex`: sets whether the primary index should be created for this bucket.

```go
bucket := NewBucket(
	"bucketName",
	WithQuota(100),
	WithReplicas(1),
	WithFlushEnabled(true),
	WithPrimaryIndex(true),
)
```

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
	`WithAnalyticsService` and `WithEventingService`. Else, it will throw an error and the container won't be created.

<!--codeinclude-->
[Docker images](../../modules/couchbase/couchbase_test.go) inside_block:dockerImages
<!--/codeinclude-->

