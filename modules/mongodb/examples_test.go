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

func ExampleRun() {
	// runMongoDBContainer {
	ctx := context.Background()

	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	defer func() {
		if err := testcontainers.TerminateContainer(mongodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mongodbContainer.State(ctx)
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

	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	defer func() {
		if err := testcontainers.TerminateContainer(mongodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	endpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
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

func ExampleRun_withCredentials() {
	ctx := context.Background()

	ctr, err := mongodb.Run(ctx,
		"mongo:6",
		mongodb.WithUsername("root"),
		mongodb.WithPassword("password"),
		testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connStr, err := ctr.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		log.Printf("failed to connect to MongoDB: %s", err)
		return
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Printf("failed to ping MongoDB: %s", err)
		return
	}
	fmt.Println(strings.Split(connStr, "@")[0])

	// Output:
	// mongodb://root:password
}
