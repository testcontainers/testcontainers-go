package minio_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcminio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
		tcminio.WithUsername("thisismyuser"), tcminio.WithPassword("thisismypassword"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	// connectionString {
	url, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(ctr.Username, ctr.Password, ""),
		Secure: false,
	})
	require.NoError(t, err)

	bucketName := "testcontainers"
	location := "eu-west-2"

	// create bucket
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	require.NoError(t, err)

	objectName := "testdata"
	contentType := "applcation/octet-stream"
	content := strings.Repeat("this is some text\n", 1048576) // (16 chars) * 1MB
	contentLength := int64(len(content))

	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectName, strings.NewReader(content), contentLength, minio.PutObjectOptions{ContentType: contentType})
	require.NoError(t, err)

	// object is a readSeekCloser
	object, err := minioClient.GetObject(ctx, uploadInfo.Bucket, uploadInfo.Key, minio.GetObjectOptions{})
	require.NoError(t, err)

	defer object.Close()

	n, err := io.Copy(io.Discard, object)
	require.NoError(t, err)
	require.Equal(t, contentLength, n)
}
