package s3mock_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/s3mock"
)

// s3EndpointResolver routes all S3 calls to the S3Mock container endpoint.
// endpointResolver {
type s3EndpointResolver struct {
	endpointURL string
}

func (r *s3EndpointResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	params.Endpoint = aws.String(r.endpointURL)
	return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// }

func newS3Client(ctx context.Context, ctr *s3mock.Container) (*s3.Client, error) {
	endpointURL, err := ctr.EndpointURL(ctx)
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolverV2 = &s3EndpointResolver{endpointURL: endpointURL}
		o.UsePathStyle = true
	})

	return client, nil
}

func TestS3Mock(t *testing.T) {
	ctx := context.Background()

	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("EndpointURL", func(t *testing.T) {
		// endpointURL {
		endpointURL, err := ctr.EndpointURL(ctx)
		// }
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(endpointURL, "http://"), "expected http scheme, got: %s", endpointURL)
	})

	t.Run("HTTPSEndpointURL", func(t *testing.T) {
		// httpsEndpointURL {
		httpsURL, err := ctr.HTTPSEndpointURL(ctx)
		// }
		require.NoError(t, err)
		// We only verify the scheme; making a full TLS request would require
		// trusting S3Mock's self-signed certificate, which is out of scope here.
		require.True(t, strings.HasPrefix(httpsURL, "https://"), "expected https scheme, got: %s", httpsURL)
	})

	t.Run("S3 operations", func(t *testing.T) {
		client, err := newS3Client(ctx, ctr)
		require.NoError(t, err)

		bucketName := "testcontainers-s3mock"

		// Create bucket
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		require.NoError(t, err)

		// Put object
		const content = "hello from testcontainers"
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
			Body:   strings.NewReader(content),
		})
		require.NoError(t, err)

		// Get object and verify content
		getOutput, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("hello.txt"),
		})
		require.NoError(t, err)
		defer getOutput.Body.Close()

		got, err := io.ReadAll(getOutput.Body)
		require.NoError(t, err)
		require.Equal(t, content, string(got))
	})
}

func TestS3MockWithInitialBuckets(t *testing.T) {
	ctx := context.Background()

	// withInitialBuckets {
	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9",
		s3mock.WithInitialBuckets("bucket1", "bucket2"),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	client, err := newS3Client(ctx, ctr)
	require.NoError(t, err)

	// Verify the pre-created buckets exist
	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	require.NoError(t, err)

	bucketNames := make([]string, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		bucketNames = append(bucketNames, aws.ToString(b.Name))
	}

	require.Contains(t, bucketNames, "bucket1")
	require.Contains(t, bucketNames, "bucket2")
}

func TestS3MockWithInitialBucketsEmpty(t *testing.T) {
	ctx := context.Background()

	// WithInitialBuckets() with no arguments should be a no-op
	ctr, err := s3mock.Run(ctx, "adobe/s3mock:3.9",
		s3mock.WithInitialBuckets(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	client, err := newS3Client(ctx, ctr)
	require.NoError(t, err)

	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	require.NoError(t, err)
	require.Empty(t, out.Buckets, "expected no pre-created buckets when WithInitialBuckets is called with no arguments")
}
