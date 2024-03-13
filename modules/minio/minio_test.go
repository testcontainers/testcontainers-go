package minio_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	container, err := tcminio.RunContainer(ctx,
		testcontainers.WithImage("minio/minio:RELEASE.2024-01-16T16-07-38Z"),
		tcminio.WithUsername("thisismyuser"), tcminio.WithPassword("thisismypassword"))
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
	// connectionString {
	url, err := container.ConnectionString(ctx)
	// }
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
	content := strings.Repeat("this is some text\n", 1048576) // (16 chars) * 1MB
	contentLength := int64(len(content))

	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectName, strings.NewReader(content), contentLength, minio.PutObjectOptions{ContentType: contentType})
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

	if n != contentLength {
		t.Fatalf("expected %d; got %d", contentLength, n)
	}
}
