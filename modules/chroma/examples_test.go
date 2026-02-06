package chroma_test

import (
	"context"
	"fmt"
	"log"
	"math"
	"path/filepath"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func ExampleRun() {
	// runChromaContainer {
	ctx := context.Background()

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:1.4.0")
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

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:1.4.0")
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
	chromaClient, err := chromago.NewHTTPClient(chromago.WithBaseURL(endpoint))
	if err != nil {
		log.Printf("failed to get client: %s", err)
		return
	}

	errHb := chromaClient.Heartbeat(context.Background())
	// }
	fmt.Println(errHb) // error is only returned if the heartbeat fails
	// closeClient {
	// ensure all resources are freed, e.g. TCP connections or the default embedding function which runs locally
	defer func() {
		err = chromaClient.Close()
		if err != nil {
			log.Printf("failed to close client: %s", err)
		}
	}()
	// }
	// Output:
	// <nil>
}

func ExampleChromaContainer_collections() {
	ctx := context.Background()

	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:1.4.0",
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      filepath.Join("testdata", "v1-config.yaml"),
			ContainerFilePath: "/config.yaml",
			FileMode:          0o644,
		}),
	)
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
	chromaClient, err := chromago.NewHTTPClient(chromago.WithBaseURL(endpoint))
	// }
	if err != nil {
		log.Printf("failed to get client: %s", err)
		return
	}
	defer func() {
		if err := chromaClient.Close(); err != nil {
			log.Printf("failed to close client: %s", err)
		}
	}()
	// reset {
	err = chromaClient.Reset(context.Background())
	// }
	if err != nil {
		log.Printf("failed to reset: %s", err)
		return
	}
	fmt.Printf("Reset successful\n")

	// createCollection {
	col, err := chromaClient.GetOrCreateCollection(context.Background(), "test-collection",
		chromago.WithCollectionMetadataCreate(
			chromago.NewMetadata(
				chromago.NewStringAttribute("str", "hello2"),
				chromago.NewIntAttribute("int", 1),
				chromago.NewFloatAttribute("float", 1.1),
			),
		),
	)
	if err != nil {
		log.Printf("failed to create collection: %s", err)
		return
	}

	fmt.Println("Collection created:", col.Name())

	// addData {
	// verify it's possible to add data to the collection
	err = col.Add(context.Background(),
		chromago.WithIDs("1", "2"),
		chromago.WithTexts("hello world", "goodbye world"),
		chromago.WithMetadatas(
			chromago.NewDocumentMetadata(chromago.NewIntAttribute("int", 1)),
			chromago.NewDocumentMetadata(chromago.NewStringAttribute("str1", "hello2")),
		))
	// }
	if err != nil {
		log.Printf("failed to add data to collection: %s", err)
		return
	}
	count, err := col.Count(context.Background())
	if err != nil {
		log.Printf("failed to count collection: %s", err)
		return
	}
	fmt.Println(count)

	// queryCollection {
	// verify it's possible to query the collection
	queryResults, err := col.Query(
		context.Background(),
		chromago.WithQueryTexts("say hello"),
		chromago.WithNResults(1),
	)
	// }
	if err != nil {
		log.Printf("failed to query collection: %s", err)
		return
	}

	distance := queryResults.GetDistancesGroups()[0][0]
	fmt.Printf("Result of query: %v %v distance_ok=%v\n", queryResults.GetIDGroups()[0][0], queryResults.GetDocumentsGroups()[0][0], math.Abs(float64(distance)-0.7525849) < 1e-3)

	// listCollections {
	cols, err := chromaClient.ListCollections(context.Background())
	// }
	if err != nil {
		log.Printf("failed to list collections: %s", err)
		return
	}

	fmt.Println(len(cols))

	// deleteCollection {
	err = chromaClient.DeleteCollection(context.Background(), "test-collection")
	// }
	if err != nil {
		log.Printf("failed to delete collection: %s", err)
		return
	}

	fmt.Println(err)

	// Output:
	// Reset successful
	// Collection created: test-collection
	// 2
	// Result of query: 1 hello world distance_ok=true
	// 1
	// <nil>
}
