package chroma_test

import (
	"context"
	"fmt"
	"log"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func ExampleRun() {
	// runChromaContainer {
	ctx := context.Background()

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:0.4.24")
	defer func() {
		if err := testcontainers.TerminateContainer(chromaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := chromaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleChromaContainer_connectWithClient() {
	// createClient {
	ctx := context.Background()

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:0.4.24")
	defer func() {
		if err := testcontainers.TerminateContainer(chromaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	endpoint, err := chromaContainer.RESTEndpoint(context.Background())
	if err != nil {
		log.Printf("failed to get REST endpoint: %s", err)
		return
	}
	chromaClient, err := chromago.NewClient(endpoint)
	if err != nil {
		log.Printf("failed to get client: %s", err)
		return
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

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:0.4.24", testcontainers.WithEnv(map[string]string{"ALLOW_RESET": "true"}))
	defer func() {
		if err := testcontainers.TerminateContainer(chromaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// getClient {
	// create the client connection and confirm that we can access the server with it
	endpoint, err := chromaContainer.RESTEndpoint(context.Background())
	if err != nil {
		log.Printf("failed to get REST endpoint: %s", err)
		return
	}
	chromaClient, err := chromago.NewClient(endpoint)
	// }
	if err != nil {
		log.Printf("failed to get client: %s", err)
		return
	}
	// reset {
	reset, err := chromaClient.Reset(context.Background())
	// }
	if err != nil {
		log.Printf("failed to reset: %s", err)
		return
	}
	fmt.Printf("Reset successful: %v\n", reset)

	// createCollection {
	// for testing we use a dummy hashing function NewConsistentHashEmbeddingFunction
	col, err := chromaClient.CreateCollection(context.Background(), "test-collection", map[string]any{}, true, types.NewConsistentHashEmbeddingFunction(), types.L2)
	// }
	if err != nil {
		log.Printf("failed to create collection: %s", err)
		return
	}

	fmt.Println("Collection created:", col.Name)

	// addData {
	// verify it's possible to add data to the collection
	col1, err := col.Add(
		context.Background(),
		nil,                                      // embeddings
		[]map[string]interface{}{},               // metadata
		[]string{"test-doc-1", "test-doc-2"},     // documents
		[]string{"test-label-1", "test-label-2"}, // ids
	)
	// }
	if err != nil {
		log.Printf("failed to add data to collection: %s", err)
		return
	}

	fmt.Println(col1.Count(context.Background()))

	// queryCollection {
	// verify it's possible to query the collection
	queryResults, err := col1.QueryWithOptions(
		context.Background(),
		types.WithQueryTexts([]string{"test-doc-1"}),
		types.WithInclude(types.IDocuments, types.IEmbeddings, types.IMetadatas),
		types.WithNResults(1),
	)
	// }
	if err != nil {
		log.Printf("failed to query collection: %s", err)
		return
	}

	fmt.Printf("Result of query: %v\n", queryResults)

	// listCollections {
	cols, err := chromaClient.ListCollections(context.Background())
	// }
	if err != nil {
		log.Printf("failed to list collections: %s", err)
		return
	}

	fmt.Println(len(cols))

	// deleteCollection {
	_, err = chromaClient.DeleteCollection(context.Background(), "test-collection")
	// }
	if err != nil {
		log.Printf("failed to delete collection: %s", err)
		return
	}

	fmt.Println(err)

	// Output:
	// Reset successful: true
	// Collection created: test-collection
	// 2 <nil>
	// Result of query: &{[[test-doc-1]] [[test-label-1]] [[map[]]] []}
	// 1
	// <nil>
}
