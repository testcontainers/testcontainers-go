package atlaslocal_test

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb/atlaslocal"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ExampleRun() {
	// runMongoDBAtlasLocalContainer {
	ctx := context.Background()

	atlaslocalContainer, err := atlaslocal.Run(ctx, "mongodb/mongodb-atlas-local:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(atlaslocalContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := atlaslocalContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	// connectToMongo {
	ctx := context.Background()

	atlaslocalContainer, err := atlaslocal.Run(ctx, "mongodb/mongodb-atlas-local:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(atlaslocalContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connString, err := atlaslocalContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		log.Printf("failed to connect to MongoDB: %s", err)
		return
	}
	// }

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Printf("failed to ping MongoDB: %s", err)
		return
	}

	fmt.Println(mongoClient.Database("test").Name())

	// Output:
	// test
}

func ExampleRun_readMongotLogs() {
	// readMongotLogs {
	ctx := context.Background()

	atlaslocalContainer, err := atlaslocal.Run(ctx, "mongodb/mongodb-atlas-local:latest",
		atlaslocal.WithMongotLogFile())

	defer func() {
		if err := testcontainers.TerminateContainer(atlaslocalContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connString, err := atlaslocalContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	_, err = mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		log.Printf("failed to connect to MongoDB: %s", err)
		return
	}

	reader, err := atlaslocalContainer.ReadMongotLogs(ctx)
	if err != nil {
		log.Printf("failed to read mongot logs: %s", err)
		return
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		log.Printf("failed to write mongot logs: %s", err)
		return
	}
	// }

	// Output:
}
