package couchbase_test

import (
	"context"
	"fmt"

	"github.com/couchbase/gocb/v2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/couchbase"
)

func ExampleRunContainer() {
	// runCouchbaseContainer {
	ctx := context.Background()

	bucketName := "testBucket"
	bucket := couchbase.NewBucket(bucketName)

	bucket = bucket.WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(false).
		WithPrimaryIndex(true)

	couchbaseContainer, err := couchbase.RunContainer(ctx,
		testcontainers.WithImage("couchbase:community-7.1.1"),
		couchbase.WithAdminCredentials("testcontainers", "testcontainers.IS.cool!"),
		couchbase.WithBuckets(bucket),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := couchbaseContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := couchbaseContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	connectionString, err := couchbaseContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: couchbaseContainer.Username(),
		Password: couchbaseContainer.Password(),
	})
	if err != nil {
		panic(err)
	}

	buckets, err := cluster.Buckets().GetAllBuckets(nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(buckets))
	fmt.Println(buckets[bucketName].Name)

	// Output:
	// true
	// 1
	// testBucket
}
