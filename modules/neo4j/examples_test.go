package neo4j_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/neo4j"
)

func ExampleRun() {
	// runNeo4jContainer {
	ctx := context.Background()

	testPassword := "letmein!"

	neo4jContainer, err := neo4j.Run(ctx,
		"docker.io/neo4j:4.4",
		neo4j.WithAdminPassword(testPassword),
		neo4j.WithLabsPlugin(neo4j.Apoc),
		neo4j.WithNeo4jSetting("dbms.tx_log.rotation.size", "42M"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := neo4jContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := neo4jContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
