package couchbase_test

import (
	"context"
	"fmt"
	"log"

	"github.com/couchbase/gocb/v2"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(couchbaseContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := couchbaseContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionString, err := couchbaseContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: couchbaseContainer.Username(),
		Password: couchbaseContainer.Password(),
	})
	if err != nil {
		log.Printf("failed to connect to cluster: %s", err)
		return
	}

	buckets, err := cluster.Buckets().GetAllBuckets(nil)
	if err != nil {
		log.Printf("failed to get buckets: %s", err)
		return
	}

	fmt.Println(len(buckets))
	fmt.Println(buckets[bucketName].Name)

	// Output:
	// true
	// 1
	// testBucket
}
