package arangodb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"

	"github.com/testcontainers/testcontainers-go"
	tcarangodb "github.com/testcontainers/testcontainers-go/modules/arangodb"
)

func ExampleRun() {
	// runArangoDBContainer {
	ctx := context.Background()

	const password = "t3stc0ntain3rs!"

	arangodbContainer, err := tcarangodb.Run(ctx, "arangodb:3.11.5", tcarangodb.WithRootPassword(password))
	defer func() {
		if err := testcontainers.TerminateContainer(arangodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := arangodbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_usingClient() {
	ctx := context.Background()

	const password = "t3stc0ntain3rs!"

	arangodbContainer, err := tcarangodb.Run(
		ctx, "arangodb:3.11.5",
		tcarangodb.WithRootPassword(password),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(arangodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	httpAddress, err := arangodbContainer.HTTPEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get transport address: %s", err)
		return
	}

	// Create an HTTP connection to the database
	endpoint := connection.NewRoundRobinEndpoints([]string{httpAddress})
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint, true))

	// Add authentication
	auth := connection.NewBasicAuth(arangodbContainer.Credentials())
	err = conn.SetAuthentication(auth)
	if err != nil {
		log.Printf("Failed to set authentication: %v", err)
		return
	}

	// Create a client
	client := arangodb.NewClient(conn)

	// Ask the version of the server
	versionInfo, err := client.Version(context.Background())
	if err != nil {
		log.Printf("Failed to get version info: %v", err)
		return
	}

	fmt.Println(versionInfo.Server)

	// Output:
	// arango
}
