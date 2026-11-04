package s3mock_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/s3mock"
)

// ExampleRun demonstrates how to start an S3Mock container with a pre-created bucket.
// See examples/ for AWS SDK client usage.
func ExampleRun() {
	// runS3MockContainer {
	ctx := context.Background()

	s3mockContainer, err := s3mock.Run(ctx,
		"adobe/s3mock:3.9.1",
		s3mock.WithInitialBuckets("mybucket"),
	)
	defer func() {
		if s3mockContainer != nil {
			if err := testcontainers.TerminateContainer(s3mockContainer); err != nil {
				log.Printf("failed to terminate container: %s", err)
			}
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := s3mockContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
