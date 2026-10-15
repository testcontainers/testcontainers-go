package garage_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/garage"
)

func ExampleRun() {
	ctx := context.Background()
	garageContainer, err := garage.Run(ctx, "dxflrs/garage:v2.4.1")
	defer func() {
		if err := testcontainers.TerminateContainer(garageContainer, testcontainers.StopTimeout(0)); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	endpoint, err := garageContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get endpoint: %s", err)
		return
	}
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       garageContainer.Region(),
		Credentials:  credentials.NewStaticCredentialsProvider(garageContainer.AccessKey(), garageContainer.SecretKey(), ""),
		UsePathStyle: true,
	})
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(garageContainer.Bucket())})
	fmt.Println(err == nil)
	// Output: true
}
