package openfga_test

import (
	"context"
	"fmt"
	"log"

	"github.com/openfga/go-sdk/client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func ExampleRunContainer() {
	// runOpenFGAContainer {
	ctx := context.Background()

	openfgaContainer, err := openfga.RunContainer(ctx, testcontainers.WithImage("openfga/openfga:v1.5.0"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := openfgaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := openfgaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connectWithSDKClient() {
	openfgaContainer, err := openfga.RunContainer(context.Background(), testcontainers.WithImage("openfga/openfga:v1.5.0"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := openfgaContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	httpEndpoint, err := openfgaContainer.HttpEndpoint(context.Background())
	if err != nil {
		log.Fatalf("failed to get HTTP endpoint: %s", err) // nolint:gocritic
	}

	// StoreId is not required for listing and creating stores
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: httpEndpoint, // required
	})
	if err != nil {
		log.Fatalf("failed to create SDK client: %s", err) // nolint:gocritic
	}

	list, err := fgaClient.ListStores(context.Background()).Execute()
	if err != nil {
		log.Fatalf("failed to list stores: %s", err) // nolint:gocritic
	}

	fmt.Println(len(list.Stores))

	store, err := fgaClient.CreateStore(context.Background()).Body(client.ClientCreateStoreRequest{Name: "test"}).Execute()
	if err != nil {
		log.Fatalf("failed to create store: %s", err) // nolint:gocritic
	}

	fmt.Println(store.Name)

	list, err = fgaClient.ListStores(context.Background()).Execute()
	if err != nil {
		log.Fatalf("failed to list stores: %s", err) // nolint:gocritic
	}

	fmt.Println(len(list.Stores))

	// Output:
	// 0
	// test
	// 1
}
