package couchbase_test

import (
	"context"
	"fmt"
	"log"

	"github.com/couchbase/gocb/v2"

	"github.com/testcontainers/testcontainers-go/modules/couchbase"
)

func ExampleRun() {
	// runCouchbaseContainer {
	ctx := context.Background()

	bucketName := "testBucket"
	bucket := couchbase.NewBucket(bucketName)

	bucket = bucket.WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(false).
		WithPrimaryIndex(true)

	couchbaseContainer, err := couchbase.Run(ctx,
		"couchbase:community-7.1.1",
		couchbase.WithAdminCredentials("testcontainers", "testcontainers.IS.cool!"),
		couchbase.WithBuckets(bucket),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := couchbaseContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := couchbaseContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	connectionString, err := couchbaseContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: couchbaseContainer.Username(),
		Password: couchbaseContainer.Password(),
	})
	if err != nil {
		log.Fatalf("failed to connect to cluster: %s", err)
	}

	buckets, err := cluster.Buckets().GetAllBuckets(nil)
	if err != nil {
		log.Fatalf("failed to get buckets: %s", err)
	}

	fmt.Println(len(buckets))
	fmt.Println(buckets[bucketName].Name)

	// Output:
	// true
	// 1
	// testBucket
}
