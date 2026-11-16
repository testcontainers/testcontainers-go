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
	endpoint, err := chromaContainer.RESTEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get REST endpoint: %s", err)
		return
	}
	chromaClient, err := chromago.NewHTTPClient(chromago.WithBaseURL(endpoint))
	if err != nil {
		log.Printf("failed to get client: %s", err)
		return
	}
	errHb := chromaClient.Heartbeat(ctx)
	// }
	fmt.Println(errHb)
	// closeClient {
	defer func() {
		if err := chromaClient.Close(); err != nil {
			log.Printf("failed to close client: %s", err)
		}
	}()
	// }
	// Output:
	// <nil>
}

func ExampleChromaContainer_collections() {
	ctx := context.Background()
	chromaContainer, err := chroma.Run(ctx, "chromadb/chroma:1.4.0", testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath: filepath.Join("..", "testdata", "v1-config.yaml"), ContainerFilePath: "/config.yaml", FileMode: 0o644,
	}))
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
	endpoint, err := chromaContainer.RESTEndpoint(ctx)
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
	if err = chromaClient.Reset(ctx); err != nil {
		log.Printf("failed to reset: %s", err)
		return
	}
	fmt.Println("Reset successful")
	// createCollection {
	col, err := chromaClient.GetOrCreateCollection(ctx, "test-collection", chromago.WithCollectionMetadataCreate(chromago.NewMetadata(chromago.NewStringAttribute("str", "hello2"), chromago.NewIntAttribute("int", 1), chromago.NewFloatAttribute("float", 1.1))))
	// }
	if err != nil {
		log.Printf("failed to create collection: %s", err)
		return
	}
	fmt.Println("Collection created:", col.Name())
	// addData {
	err = col.Add(ctx, chromago.WithIDs("1", "2"), chromago.WithTexts("hello world", "goodbye world"), chromago.WithMetadatas(chromago.NewDocumentMetadata(chromago.NewIntAttribute("int", 1)), chromago.NewDocumentMetadata(chromago.NewStringAttribute("str1", "hello2"))))
	// }
	if err != nil {
		log.Printf("failed to add data: %s", err)
		return
	}
	count, err := col.Count(ctx)
	if err != nil {
		log.Printf("failed to count collection: %s", err)
		return
	}
	fmt.Println(count)
	// queryCollection {
	queryResults, err := col.Query(ctx, chromago.WithQueryTexts("say hello"), chromago.WithNResults(1))
	// }
	if err != nil {
		log.Printf("failed to query collection: %s", err)
		return
	}
	distance := queryResults.GetDistancesGroups()[0][0]
	fmt.Printf("Result of query: %v %v distance_ok=%v\n", queryResults.GetIDGroups()[0][0], queryResults.GetDocumentsGroups()[0][0], math.Abs(float64(distance)-0.7525849) < 1e-3)
	// listCollections {
	cols, err := chromaClient.ListCollections(ctx)
	// }
	if err != nil {
		log.Printf("failed to list collections: %s", err)
		return
	}
	fmt.Println(len(cols))
	// deleteCollection {
	err = chromaClient.DeleteCollection(ctx, "test-collection")
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
