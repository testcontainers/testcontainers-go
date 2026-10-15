package garage_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/garage"
)

func TestGarage(t *testing.T) {
	for _, tt := range []struct {
		name                                 string
		image                                string
		opts                                 []testcontainers.ContainerCustomizer
		accessKey, secretKey, bucket, region string
	}{
		{name: "minimum version", image: "dxflrs/garage:v2.3.0", accessKey: "testcontainers", secretKey: "testcontainers-secret", bucket: "testcontainers", region: "garage"},
		{name: "defaults", image: "dxflrs/garage:v2.4.1", accessKey: "testcontainers", secretKey: "testcontainers-secret", bucket: "testcontainers", region: "garage"},
		{name: "custom", image: "dxflrs/garage:v2.4.1", opts: []testcontainers.ContainerCustomizer{garage.WithCredentials("custom-access", "custom-secret-key"), garage.WithBucket("custom-bucket"), garage.WithRegion("eu-west-1")}, accessKey: "custom-access", secretKey: "custom-secret-key", bucket: "custom-bucket", region: "eu-west-1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctr, err := garage.Run(ctx, tt.image, tt.opts...)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.NoError(t, err)
			require.Equal(t, tt.accessKey, ctr.AccessKey())
			require.Equal(t, tt.secretKey, ctr.SecretKey())
			require.Equal(t, tt.bucket, ctr.Bucket())
			require.Equal(t, tt.region, ctr.Region())
			endpoint, err := ctr.ConnectionString(ctx)
			require.NoError(t, err)
			client := s3.New(s3.Options{
				BaseEndpoint: aws.String(endpoint), Region: ctr.Region(), UsePathStyle: true,
				Credentials: credentials.NewStaticCredentialsProvider(ctr.AccessKey(), ctr.SecretKey(), ""),
			})
			_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(ctr.Bucket())})
			require.NoError(t, err)
			_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("another-bucket"), CreateBucketConfiguration: &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraint(ctr.Region())}})
			require.NoError(t, err)
			for _, bucket := range []string{ctr.Bucket(), "another-bucket"} {
				_, err = client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String("hello.txt"), Body: strings.NewReader("hello Garage")})
				require.NoError(t, err)
				object, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String("hello.txt")})
				require.NoError(t, err)
				content, err := io.ReadAll(object.Body)
				require.NoError(t, object.Body.Close())
				require.NoError(t, err)
				require.Equal(t, "hello Garage", string(content))
				listed, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
				require.NoError(t, err)
				require.Len(t, listed.Contents, 1)
				require.Equal(t, "hello.txt", aws.ToString(listed.Contents[0].Key))
				signed, err := s3.NewPresignClient(client).PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String("hello.txt")})
				require.NoError(t, err)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, signed.URL, nil)
				require.NoError(t, err)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				content, err = io.ReadAll(resp.Body)
				require.NoError(t, resp.Body.Close())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "hello Garage", string(content))
				_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String("hello.txt")})
				require.NoError(t, err)
				listed, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
				require.NoError(t, err)
				require.Empty(t, listed.Contents)
			}
		})
	}
}

func TestInvalidOptions(t *testing.T) {
	for _, tt := range []struct {
		name string
		opt  testcontainers.ContainerCustomizer
		want string
	}{
		{"access key", garage.WithCredentials("short", "valid-secret-key"), "access key must be at least 8 bytes"},
		{"secret key", garage.WithCredentials("valid-access", "short"), "secret key must be at least 16 bytes"},
		{"bucket", garage.WithBucket(""), "bucket must not be empty"},
		{"region", garage.WithRegion(""), "region must not be empty"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctr, err := garage.Run(context.Background(), "dxflrs/garage:v2.4.1", tt.opt)
			testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
			require.ErrorContains(t, err, tt.want)
			require.Nil(t, ctr)
		})
	}
}
