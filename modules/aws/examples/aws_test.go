package examples_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	awsemulator "github.com/testcontainers/testcontainers-go/modules/aws"
)

func TestAWSServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := awsemulator.Run(ctx, "ghcr.io/xsavikx/aws-emulator:1.0.509.1-20261118.43", awsemulator.WithDisableTelemetry())
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	// awsSDKConfig {
	endpoint, err := container.EndpointURL(ctx)
	require.NoError(t, err)
	keys, err := container.Credentials(ctx)
	require.NoError(t, err)
	tlsConfig, err := container.TLSConfig(ctx)
	require.NoError(t, err)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	t.Cleanup(transport.CloseIdleConnections)
	cfg := aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(keys.AccessKeyID, keys.SecretAccessKey, ""),
		HTTPClient:   &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}
	// }

	t.Run("STS", func(t *testing.T) {
		client := sts.NewFromConfig(cfg)
		identity, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		require.NoError(t, err)
		require.Equal(t, "000000000000", aws.ToString(identity.Account))
	})

	t.Run("invalid credentials rejected", func(t *testing.T) {
		client := sts.NewFromConfig(cfg, func(o *sts.Options) {
			o.Credentials = credentials.NewStaticCredentialsProvider("invalid", "invalid", "")
		})
		_, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		require.Error(t, err)
	})

	t.Run("S3", func(t *testing.T) {
		client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })
		bucket := aws.String("testcontainers-prototype")
		key := aws.String("hello.txt")
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: bucket})
		require.NoError(t, err)
		const content = "hello from testcontainers"
		_, err = client.PutObject(ctx, &s3.PutObjectInput{Bucket: bucket, Key: key, Body: strings.NewReader(content)})
		require.NoError(t, err)
		obj, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: bucket, Key: key})
		require.NoError(t, err)
		defer obj.Body.Close()
		data, err := io.ReadAll(obj.Body)
		require.NoError(t, err)
		require.Equal(t, content, string(data))
	})

	t.Run("SQS", func(t *testing.T) {
		client := sqs.NewFromConfig(cfg)
		queue, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String("testcontainers-prototype")})
		require.NoError(t, err)
		// Keep QueueUrl as the opaque identifier returned by the emulator.
		// BaseEndpoint controls the actual SDK connection through the mapped port.
		_, err = client.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: queue.QueueUrl, MessageBody: aws.String("hello")})
		require.NoError(t, err)
		messages, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: queue.QueueUrl, WaitTimeSeconds: 1})
		require.NoError(t, err)
		require.Len(t, messages.Messages, 1)
		require.Equal(t, "hello", aws.ToString(messages.Messages[0].Body))
		_, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{QueueUrl: queue.QueueUrl, ReceiptHandle: messages.Messages[0].ReceiptHandle})
		require.NoError(t, err)
	})
}
