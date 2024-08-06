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
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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

	container, err := localstack.Run(ctx, "localstack/localstack:1.4.0")
	assert.NilError(t, err)

	s3Client, err := s3Client(ctx, container)
	assert.NilError(t, err)

	t.Run("S3 operations", func(t *testing.T) {
		bucketName := "localstack-bucket"

		// Create Bucket
		outputBucket, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		assert.NilError(t, err)
		assert.Check(t, outputBucket != nil)

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
		assert.NilError(t, err)
		assert.Check(t, outputObject != nil)

		t.Run("List Buckets", func(t *testing.T) {
			output, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
			assert.NilError(t, err)
			assert.Check(t, output != nil)

			buckets := output.Buckets
			assert.Check(t, is.Len(buckets, 1))
			assert.Check(t, is.Equal(bucketName, *buckets[0].Name))
		})

		t.Run("List Objects in Bucket", func(t *testing.T) {
			output, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket: aws.String(bucketName),
			})
			assert.NilError(t, err)
			assert.Check(t, output != nil)

			objects := output.Contents

			assert.Check(t, is.Len(objects, 1))
			assert.Check(t, is.Equal(s3Key1, *objects[0].Key))
			assert.Check(t, is.DeepEqual(aws.Int64(int64(len(body1))), objects[0].Size))
		})
	})
}
