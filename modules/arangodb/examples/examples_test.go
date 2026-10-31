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

func ExampleRun_usingClient() {
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

	httpAddress, err := arangodbContainer.HTTPEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get transport address: %s", err)
		return
	}
	endpoint := connection.NewRoundRobinEndpoints([]string{httpAddress})
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint, true))
	if err := conn.SetAuthentication(connection.NewBasicAuth(arangodbContainer.Credentials())); err != nil {
		log.Printf("failed to set authentication: %v", err)
		return
	}
	versionInfo, err := arangodb.NewClient(conn).Version(ctx)
	if err != nil {
		log.Printf("failed to get version info: %v", err)
		return
	}
	fmt.Println(versionInfo.Server)

	// Output:
	// arango
}
