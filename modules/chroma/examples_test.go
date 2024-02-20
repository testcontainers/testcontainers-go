package chroma_test

import (
	"context"
	"fmt"
	"log"
	"os"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	chromaopenapi "github.com/amikos-tech/chroma-go/swagger"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func ExampleRunContainer() {
	// runChromaContainer {
	ctx := context.Background()

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22"))
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

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22"))
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

func ExampleChromaContainer_collections() {
	ctx := context.Background()

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

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

	// createCollection {
	// for testing purposes, the OPENAI_API_KEY environment variable can be empty
	// therefore this test is expected to succeed even though the API key is not set.
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(os.Getenv("OPENAI_API_KEY"))
	distanceFunction := chromago.L2

	col, err := chromaClient.CreateCollection(context.Background(), "test-collection", map[string]any{}, true, embeddingFunction, distanceFunction)
	// }
	if err != nil {
		log.Fatalf("failed to create collection: %s", err) // nolint:gocritic
	}

	fmt.Println("Collection created:", col.Name)

	// verify it's possible to add data to the collection
	col1, err := col.Add(
		context.Background(),
		[][]float32{{1, 2, 3}, {4, 5, 6}},
		[]map[string]interface{}{},
		[]string{"test-doc-1", "test-doc-2"},
		[]string{"test-label-1", "test-label-2"},
	)
	if err != nil {
		log.Fatalf("failed to add data to collection: %s", err) // nolint:gocritic
	}

	fmt.Println(col1.Count(context.Background()))

	// listCollections {
	cols, err := chromaClient.ListCollections(context.Background())
	// }
	if err != nil {
		log.Fatalf("failed to list collections: %s", err) // nolint:gocritic
	}

	fmt.Println(len(cols))

	// deleteCollection {
	_, err = chromaClient.DeleteCollection(context.Background(), "test-collection")
	// }
	if err != nil {
		log.Fatalf("failed to delete collection: %s", err) // nolint:gocritic
	}

	fmt.Println(err)

	// Output:
	// Collection created: test-collection
	// 2 <nil>
	// 1
	// <nil>
}
