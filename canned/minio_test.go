package canned

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestMakeBucketInMinio(t *testing.T) {

	ctx := context.Background()

	c, err := NewMinioContainer(ctx, MinioContainerRequest{
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

func ExampleMinioContainerRequest() {

	// Optional
	containerRequest := testcontainers.ContainerRequest{
		Image: "docker.io/minio/minio:latest",
	}

	// Required if autostart is needed
	genericContainerRequest := testcontainers.GenericContainerRequest{
		Started:          true,
		ContainerRequest: containerRequest,
	}

	// AccessKey and SecretKey are optionmal
	minioContainerRequest := MinioContainerRequest{
		AccessKey:               "AKIAIOSFODNN7EXAMPLE",
		SecretKey:               "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		GenericContainerRequest: genericContainerRequest,
	}

	minioContainerRequest.Validate()
}

func ExampleNewMinioContainer() {
	ctx := context.Background()

	c, _ := NewMinioContainer(ctx, MinioContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})

	defer c.Container.Terminate(ctx)
}

func ExampleMinioContainer_GetClient() {
	ctx := context.Background()

	c, _ := NewMinioContainer(ctx, MinioContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			Started: true,
		},
	})

	minioClient, _ := c.GetClient(ctx)

	minioClient.ListBuckets()
}
