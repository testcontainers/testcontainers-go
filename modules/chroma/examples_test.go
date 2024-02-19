package chroma_test

import (
	"context"
	"fmt"
	"log"

	chromago "github.com/amikos-tech/chroma-go"
	chromaopenapi "github.com/amikos-tech/chroma-go/swagger"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func ExampleRunContainer() {
	// runChromaContainer {
	ctx := context.Background()

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22.dev44"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := chromaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := chromaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleChromaContainer_connectWithClient() {
	// createClient {
	ctx := context.Background()

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22.dev44"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := chromaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	connectionStr, err := chromaContainer.RESTEndpoint(ctx)
	if err != nil {
		log.Fatalf("failed to get REST endpoint: %s", err) // nolint:gocritic
	}

	// create the client connection and confirm that we can access the server with it
	configuration := chromaopenapi.NewConfiguration()
	configuration.Servers = chromaopenapi.ServerConfigurations{
		{
			URL:         connectionStr,
			Description: "Chromadb server url for this store",
		},
	}
	chromaClient := &chromago.Client{
		ApiClient: chromaopenapi.NewAPIClient(configuration),
	}

	hbs, errHb := chromaClient.Heartbeat(context.Background())
	// }
	if _, ok := hbs["nanosecond heartbeat"]; ok {
		fmt.Println(ok)
	}

	fmt.Println(errHb)

	// Output:
	// true
	// <nil>
}
