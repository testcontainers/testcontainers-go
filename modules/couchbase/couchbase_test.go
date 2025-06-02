package couchbase_test

import (
	"context"
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tccouchbase "github.com/testcontainers/testcontainers-go/modules/couchbase"
)

const (
	// dockerImages {
	enterpriseEdition = "couchbase:enterprise-7.6.1"
	communityEdition  = "couchbase:community-7.1.1"
	// }
)

func TestCouchbaseWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	// withBucket {
	bucketName := "testBucket"
	bucket := tccouchbase.NewBucket(bucketName)

	bucket = bucket.WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(false).
		WithPrimaryIndex(true)

	ctr, err := tccouchbase.Run(ctx, communityEdition, tccouchbase.WithBuckets(bucket))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	// }

	cluster, err := connectCluster(ctx, ctr)
	require.NoError(t, err)

	testBucketUsage(t, cluster.Bucket(bucketName))
}

func TestCouchbaseWithEnterpriseContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	bucket := tccouchbase.NewBucket(bucketName).
		WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(true).
		WithPrimaryIndex(true)
	ctr, err := tccouchbase.Run(ctx,
		enterpriseEdition,
		tccouchbase.WithBuckets(bucket),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	cluster, err := connectCluster(ctx, ctr)
	require.NoError(t, err)

	testBucketUsage(t, cluster.Bucket(bucketName))
}

func TestCouchbaseWithReuse(t *testing.T) {
	ctx := context.Background()

	containerName := "couchbase-" + testcontainers.SessionID()

	bucketName := "testBucket"
	bucket := tccouchbase.NewBucket(bucketName).
		WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(true).
		WithPrimaryIndex(true)
	ctr, err := tccouchbase.Run(ctx,
		enterpriseEdition,
		tccouchbase.WithBuckets(bucket),
		testcontainers.WithReuseByName(containerName),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	cluster, err := connectCluster(ctx, ctr)
	require.NoError(t, err)

	testBucketUsage(t, cluster.Bucket(bucketName))

	// Test reuse when first container has had time to fully start up and be configured with auth
	// Without enabling auth on the initCluster functions the reuse of this container fails with
	// "init cluster: context deadline exceeded".
	// This is due to the management endpoints requiring the Basic Auth headers once configureAdminUser
	// has completed.
	reusedCtr, err := tccouchbase.Run(ctx,
		enterpriseEdition,
		testcontainers.WithReuseByName(containerName),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.Equal(t, ctr.GetContainerID(), reusedCtr.GetContainerID())

	cluster, err = connectCluster(ctx, reusedCtr)
	require.NoError(t, err)

	testBucketUsage(t, cluster.Bucket(bucketName))
}

func TestWithCredentials(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	ctr, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithAdminCredentials("testcontainers", "testcontainers.IS.cool!"),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestWithCredentials_Password_LessThan_6(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	ctr, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithAdminCredentials("testcontainers", "12345"),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func TestAnalyticsServiceWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	ctr, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithServiceAnalytics(),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func TestEventingServiceWithCommunityContainer(t *testing.T) {
	ctx := context.Background()

	bucketName := "testBucket"
	ctr, err := tccouchbase.Run(ctx,
		communityEdition,
		tccouchbase.WithServiceEventing(),
		tccouchbase.WithBuckets(tccouchbase.NewBucket(bucketName)))
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
}

func testBucketUsage(t *testing.T, bucket *gocb.Bucket) {
	t.Helper()
	err := bucket.WaitUntilReady(5*time.Second, nil)
	require.NoErrorf(t, err, "could not connect bucket")

	key := "foo"
	data := map[string]string{"key": "value"}
	collection := bucket.DefaultCollection()

	_, err = collection.Upsert(key, data, nil)
	require.NoErrorf(t, err, "could not upsert data")

	result, err := collection.Get(key, nil)
	require.NoErrorf(t, err, "could not get data")

	var resultData map[string]string
	err = result.Content(&resultData)
	require.NoErrorf(t, err, "could not assign content")
	require.Contains(t, resultData, "key")
	require.Equalf(t, "value", resultData["key"], "Expected value to be [%s], got %s", "value", resultData["key"])
}

func connectCluster(ctx context.Context, container *tccouchbase.CouchbaseContainer) (*gocb.Cluster, error) {
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: container.Username(),
		Password: container.Password(),
	})
}
