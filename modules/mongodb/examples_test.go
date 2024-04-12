package mongodb_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRunContainer() {
	// runMongoDBContainer {
	ctx := context.Background()

	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := mongodbContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	// connectToMongo {
	ctx := context.Background()

	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	endpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %s", err)
	}
	// }

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("failed to ping MongoDB: %s", err)
	}

	fmt.Println(mongoClient.Database("test").Name())

	// Output:
	// test
}

func ExampleRunContainer_withCredentials() {
	ctx := context.Background()

	container, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:6"),
		mongodb.WithUsername("root"),
		mongodb.WithPassword("password"),
		testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %s", err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("failed to ping MongoDB: %s", err)
	}
	fmt.Println(strings.Split(connStr, "@")[0])

	// Output:
	// mongodb://root:password
}
