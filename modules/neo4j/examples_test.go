package neo4j_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(neo4jContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := neo4jContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
