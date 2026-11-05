package mongodb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func ExampleRun() {
	// runMongoDBContainer {
	ctx := context.Background()

	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	defer func() {
		if err := testcontainers.TerminateContainer(mongodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mongodbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
