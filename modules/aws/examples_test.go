package aws_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/aws"
)

func ExampleRun() {
	// runAWSContainer {
	ctx := context.Background()
	container, err := aws.Run(ctx, "ghcr.io/xsavikx/aws-emulator:1.0.509.1-20261118.43", aws.WithDisableTelemetry())
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate AWS Emulator: %s", err)
		}
	}()
	if err != nil {
		log.Printf("start AWS Emulator: %s", err)
		return
	}
	// }
	endpoint, err := container.EndpointURL(ctx)
	if err != nil {
		log.Printf("get endpoint: %s", err)
		return
	}
	fmt.Println(endpoint != "")
	// Output: true
}
