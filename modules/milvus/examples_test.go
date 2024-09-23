package milvus_test

import (
	"context"
	"fmt"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func ExampleRun() {
	// runMilvusContainer {
	ctx := context.Background()

	milvusContainer, err := milvus.Run(ctx, "milvusdb/milvus:v2.3.9")
	defer func() {
		if err := testcontainers.TerminateContainer(milvusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := milvusContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleMilvusContainer_collections() {
	// createCollections {
	ctx := context.Background()

	milvusContainer, err := milvus.Run(ctx, "milvusdb/milvus:v2.3.9")
	defer func() {
		if err := testcontainers.TerminateContainer(milvusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connectionStr, err := milvusContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	// Create a client to interact with the Milvus container
	milvusClient, err := client.NewGrpcClient(context.Background(), connectionStr)
	if err != nil {
		log.Print("failed to connect to Milvus:", err.Error())
		return
	}
	defer milvusClient.Close()

	collectionName := "book"
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Test book search",
		Fields: []*entity.Field{
			{
				Name:       "book_id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
			},
			{
				Name:       "word_count",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "book_intro",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": "2",
				},
			},
		},
		EnableDynamicField: true,
	}

	err = milvusClient.CreateCollection(
		context.Background(), // ctx
		schema,
		2, // shardNum
	)
	if err != nil {
		log.Printf("failed to create collection: %s", err)
		return
	}

	list, err := milvusClient.ListCollections(context.Background())
	if err != nil {
		log.Printf("failed to list collections: %s", err)
		return
	}
	// }

	fmt.Println(len(list))

	// Output:
	// 1
}
