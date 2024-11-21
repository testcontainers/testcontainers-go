package dynamodb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

func ExampleRun() {
	// runDynamoDBContainer {
	ctx := context.Background()

	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.1")
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run dynamodb container: %s", err)
		return
	}
	// }

	state, err := ctr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
