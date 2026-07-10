package s3mock_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/s3mock"
)

func ExampleRun() {
	// runS3MockContainer {
	ctx := context.Background()

	s3mockContainer, err := s3mock.Run(ctx,
		"adobe/s3mock:3.9",
		s3mock.WithInitialBuckets("mybucket"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(s3mockContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// awsClientSetup {
	endpointURL, err := s3mockContainer.EndpointURL(ctx)
	if err != nil {
		log.Printf("failed to get endpoint URL: %s", err)
		return
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		log.Printf("failed to load AWS config: %s", err)
		return
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolverV2 = &s3EndpointResolver{endpointURL: endpointURL}
		o.UsePathStyle = true
	})
	// }

	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("failed to list buckets: %s", err)
		return
	}

	fmt.Println(len(out.Buckets))

	// Output:
	// 1
}
