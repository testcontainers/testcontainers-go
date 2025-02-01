package pinecone_test

import (
	"context"
	"fmt"
	"log"

	"github.com/pinecone-io/go-pinecone/v2/pinecone"

	"github.com/testcontainers/testcontainers-go"
	tcpinecone "github.com/testcontainers/testcontainers-go/modules/pinecone"
)

func ExampleRun() {
	ctx := context.Background()

	pineconeContainer, err := tcpinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:v0.7.0")
	defer func() {
		if err := testcontainers.TerminateContainer(pineconeContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := pineconeContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// httpConnection {
	host, err := pineconeContainer.HttpEndpoint()
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}
	// }

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: "testcontainers-go", // API key is required, else use headers
		Host:   host,
	})
	if err != nil {
		log.Printf("failed to create pinecone client: %s", err)
		return
	}

	indexName := "my-serverless-index"

	idx, err := pc.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      indexName,
		Dimension: 3,
		Metric:    pinecone.Cosine,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1",
		Tags:      &pinecone.IndexTags{"environment": "development"},
	})
	if err != nil {
		log.Printf("failed to create serverless index: %s", err)
		return
	}

	fmt.Println(idx.Name)

	// Output:
	// true
	// my-serverless-index
}
