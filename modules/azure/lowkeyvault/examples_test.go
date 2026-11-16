package lowkeyvault_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/lowkeyvault"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal")
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := lowkeyVaultContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_secretOperationsNetwork() {
	ctx := context.Background()

	aNetwork, err := network.New(ctx)
	defer func() {
		err := aNetwork.Remove(ctx)
		if err != nil {
			log.Printf("failed to remove network: %s", err)
			return
		}
	}()
	if err != nil {
		log.Printf("failed to setup network: %s", err)
		return
	}

	// createContainerWithNetwork {
	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal",
		lowkeyvault.WithNetworkAlias("lowkey-vault", aNetwork),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// obtainEndpointUrls {
	connURL, err := lowkeyVaultContainer.ConnectionURL(ctx, lowkeyvault.Network)
	if err != nil {
		log.Printf("failed to get connection url: %s", err)
		return
	}

	identityEndpoint, err := lowkeyVaultContainer.IdentityEndpoint(ctx, lowkeyvault.Network)
	if err != nil {
		log.Printf("failed to get token url: %s", err)
		return
	}
	// }

	networkContainer, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(
			testcontainers.FromDockerfile{
				Context:    "testdata",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
		// configureClient {
		testcontainers.WithEnv(map[string]string{
			"IDENTITY_ENDPOINT": identityEndpoint,
			"IDENTITY_HEADER":   lowkeyVaultContainer.IdentityHeader(),
			"CONNECTION_URL":    connURL,
		}),
		// }
		network.WithNetwork(nil, aNetwork),
		testcontainers.WithWaitStrategy(
			wait.ForLog("true"),
		),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(networkContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	state, err := networkContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.ExitCode == 0)

	// Output:
	// true
}
