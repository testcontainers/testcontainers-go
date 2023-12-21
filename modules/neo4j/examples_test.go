package neo4j_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/neo4j"
)

func ExampleRunContainer() {
	// runNeo4jContainer {
	ctx := context.Background()

	testPassword := "letmein!"

	neo4jContainer, err := neo4j.RunContainer(ctx,
		testcontainers.WithImage("docker.io/neo4j:4.4"),
		neo4j.WithAdminPassword(testPassword),
		neo4j.WithLabsPlugin(neo4j.Apoc),
		neo4j.WithNeo4jSetting("dbms.tx_log.rotation.size", "42M"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := neo4jContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := neo4jContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
