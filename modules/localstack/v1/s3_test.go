package v1_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

// awsSDKClientV1 {
// awsSession returns a new AWS session for the given service. To retrieve the specific AWS service client, use the
// session's client method, e.g. s3manager.NewUploader(session).
func awsSession(ctx context.Context, l *localstack.LocalStackContainer) (*session.Session, error) {
	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return &session.Session{}, err
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return &session.Session{}, err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return &session.Session{}, err
	}

	awsConfig := &aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   credentials.NewStaticCredentials(accesskey, secretkey, token),
		S3ForcePathStyle:              aws.Bool(true),
		Endpoint:                      aws.String(fmt.Sprintf("http://%s:%d", host, mappedPort.Int())),
	}

	return session.NewSession(awsConfig)
}

// }

func TestS3(t *testing.T) {
	ctx := context.Background()

	ctr, err := localstack.Run(ctx, "localstack/localstack:1.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	session, err := awsSession(ctx, ctr)
	require.NoError(t, err)

	s3Uploader := s3manager.NewUploader(session)

	t.Run("S3 operations", func(t *testing.T) {
		bucketName := "localstack-bucket"

		// Create an Amazon S3 service client
		s3API := s3Uploader.S3

		// Create Bucket
		outputBucket, err := s3API.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)
		assert.NotNil(t, outputBucket)

		// put object
		s3Key1 := "key1"
		body1 := []byte("Hello from localstack 1")
		outputObject, err := s3API.PutObject(&s3.PutObjectInput{
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
			output, err := s3API.ListBuckets(nil)
			require.NoError(t, err)
			assert.NotNil(t, output)

			buckets := output.Buckets
			assert.Len(t, buckets, 1)
			assert.Equal(t, bucketName, *buckets[0].Name)
		})

		t.Run("List Objects in Bucket", func(t *testing.T) {
			output, err := s3API.ListObjects(&s3.ListObjectsInput{
				Bucket: aws.String(bucketName),
			})
			require.NoError(t, err)
			assert.NotNil(t, output)

			objects := output.Contents

			assert.Len(t, objects, 1)
			assert.Equal(t, s3Key1, *objects[0].Key)
			assert.Equal(t, int64(len(body1)), *objects[0].Size)
		})
	})
}
