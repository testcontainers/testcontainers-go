package v2_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

const (
	accesskey = "a"
	secretkey = "b"
	token     = "c"
	region    = "us-east-1"
)

// awsSDKClientV2 {
func s3Client(ctx context.Context, l *localstack.LocalStackContainer) (*s3.Client, error) {
	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return nil, err
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return nil, err
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           fmt.Sprintf("http://%s:%d", host, mappedPort.Int()),
				SigningRegion: region,
			}, nil
		})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accesskey, secretkey, token)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

// }

func TestS3(t *testing.T) {
	ctx := context.Background()

	ctr, err := localstack.Run(ctx, "localstack/localstack:1.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	s3Client, err := s3Client(ctx, ctr)
	require.NoError(t, err)

	t.Run("S3 operations", func(t *testing.T) {
		bucketName := "localstack-bucket"

		// Create Bucket
		outputBucket, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)
		assert.NotNil(t, outputBucket)

		// put object
		s3Key1 := "key1"
		body1 := []byte("Hello from localstack 1")
		outputObject, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:             aws.String(bucketName),
			Key:                aws.String(s3Key1),
			Body:               bytes.NewReader(body1),
			ContentLength:      aws.Int64(int64(len(body1))),
			ContentType:        aws.String("application/text"),
			ContentDisposition: aws.String("attachment"),
		})
		require.NoError(t, err)
		assert.NotNil(t, outputObject)

		t.Run("List Buckets", func(t *testing.T) {
			output, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
			require.NoError(t, err)
			assert.NotNil(t, output)

			buckets := output.Buckets
			assert.Len(t, buckets, 1)
			assert.Equal(t, bucketName, *buckets[0].Name)
		})

		t.Run("List Objects in Bucket", func(t *testing.T) {
			output, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket: aws.String(bucketName),
			})
			require.NoError(t, err)
			assert.NotNil(t, output)

			objects := output.Contents

			assert.Len(t, objects, 1)
			assert.Equal(t, s3Key1, *objects[0].Key)
			assert.Equal(t, aws.Int64(int64(len(body1))), objects[0].Size)
		})
	})
}
