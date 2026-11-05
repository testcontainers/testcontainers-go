package couchbase_test

import (
	"context"
	"testing"

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
