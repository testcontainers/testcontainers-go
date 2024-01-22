package minio

import (
	"context"
	"io"
	"testing"

	"github.com/adoublef/sdk/bytest"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("minio/minio:RELEASE.2024-01-16T16-07-38Z"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	url, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(container.Username, container.Password, ""),
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	bucketName := "testcontainers"
	location := "eu-west-2"

	// create bucket
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		t.Fatal(err)
	}

	objectName := "testdata"
	contentType := "applcation/octet-stream"

	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectName, bytest.NewReader(bytest.MB*16), (bytest.MB * 16), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		t.Fatal(err)
	}

	// object is a readSeekCloser
	object, err := minioClient.GetObject(ctx, uploadInfo.Bucket, uploadInfo.Key, minio.GetObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()

	n, err := io.Copy(io.Discard, object)
	if err != nil {
		t.Fatal(err)
	}

	if n != bytest.MB*16 {
		t.Fatalf("expected %d; got %d", bytest.MB*16, n)
	}
}
