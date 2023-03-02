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
```go
container, err := couchbase.StartContainer(ctx, 
	WithImageName("couchbase:community-7.1.1"), 
	WithBucket(NewBucket(bucketName)))
```
2. The **ConnectionString** method returns the connection string to connect to the Couchbase container instance. 
It returns a string with the format `couchbase://<host>:<port>`.
The **Username** method returns the username of the Couchbase administrator. 
The **Password** method returns the password of the Couchbase administrator.
```go
connectionString, err := container.ConnectionString(ctx)
if err != nil {
	return nil, err
}

cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
	Username: container.Username(),
	Password: container.Password(),
})
```