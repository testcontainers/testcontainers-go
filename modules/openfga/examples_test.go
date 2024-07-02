package openfga_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func ExampleRun() {
	// runOpenFGAContainer {
	ctx := context.Background()

	openfgaContainer, err := openfga.Run(ctx, "openfga/openfga:v1.5.0")
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

func ExampleRun_connectToPlayground() {
	openfgaContainer, err := openfga.Run(context.Background(), "openfga/openfga:v1.5.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := openfgaContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// playgroundEndpoint {
	playgroundEndpoint, err := openfgaContainer.PlaygroundEndpoint(context.Background())
	if err != nil {
		log.Fatalf("failed to get playground endpoint: %s", err) // nolint:gocritic
	}
	// }

	httpClient := http.Client{}

	resp, err := httpClient.Get(playgroundEndpoint)
	if err != nil {
		log.Fatalf("failed to get playground endpoint: %s", err) // nolint:gocritic
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

func ExampleRun_connectWithSDKClient() {
	openfgaContainer, err := openfga.Run(context.Background(), "openfga/openfga:v1.5.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := openfgaContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// httpEndpoint {
	httpEndpoint, err := openfgaContainer.HttpEndpoint(context.Background())
	if err != nil {
		log.Fatalf("failed to get HTTP endpoint: %s", err) // nolint:gocritic
	}
	// }

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

func ExampleRun_writeModel() {
	// openFGAwriteModel {
	secret := "openfga-secret"
	openfgaContainer, err := openfga.Run(
		context.Background(),
		"openfga/openfga:v1.5.0",
		testcontainers.WithEnv(map[string]string{
			"OPENFGA_LOG_LEVEL":            "warn",
			"OPENFGA_AUTHN_METHOD":         "preshared",
			"OPENFGA_AUTHN_PRESHARED_KEYS": secret,
		}),
	)
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

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: httpEndpoint,
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: secret,
			},
		},
		// because we are going to write an authorization model,
		// we need to specify an store id. Else, it will fail with
		// "Configuration.StoreId is required and must be specified to call this method"
		// In this example, it's an arbitrary store id, that will be created
		// on the fly.
		StoreId: "11111111111111111111111111",
	})
	if err != nil {
		log.Fatalf("failed to create openfga client: %v", err)
	}

	f, err := os.Open(filepath.Join("testdata", "authorization_model.json"))
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var body client.ClientWriteAuthorizationModelRequest
	if err := json.Unmarshal(bs, &body); err != nil {
		log.Fatalf("failed to unmarshal json: %v", err)
	}

	resp, err := fgaClient.WriteAuthorizationModel(context.Background()).Body(body).Execute()
	if err != nil {
		log.Fatalf("failed to write authorization model: %v", err)
	}

	// }

	value, ok := resp.GetAuthorizationModelIdOk()
	fmt.Println(ok)
	fmt.Println(*value != "")

	// Output:
	// true
	// true
}
