package mongodb_test

import (
	"context"
	"fmt"
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
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := mongodbContainer.State(ctx)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	endpoint, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		panic(err)
	}
	// }

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		panic(err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Split(connStr, "@")[0])

	// Output:
	// mongodb://root:password
}
