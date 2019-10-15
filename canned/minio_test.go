package canned

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestMakeBucketInMinio(t *testing.T) {

	ctx := context.Background()

	c, err := MinioContainer(ctx, MinioContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Container.Terminate(ctx)

	minioClient, err := c.GetClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = minioClient.MakeBucket("a-test-bucket", "us-east-1")
	if err != nil {
		t.Fatalf("error while creating bucket: %v", err)
	}
}
