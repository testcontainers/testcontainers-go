package vault_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func ExampleRun() {
	// runVaultContainer {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withToken() {
	// runVaultContainerWithToken {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0", vault.WithToken("MyToKeN"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	cmds := []string{
		"vault", "kv", "put", "secret/test", "value=123",
	}
	exitCode, _, err := vaultContainer.Exec(ctx, cmds, exec.Multiplexed())
	if err != nil {
		log.Fatalf("failed to execute command: %s", err)
	}

	fmt.Println(exitCode)

	// Output:
	// true
	// 0
}

func ExampleRun_withInitCommand() {
	// runVaultContainerWithInitCommand {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0", vault.WithToken("MyToKeN"), vault.WithInitCommand(
		"auth enable approle",                         // Enable the approle auth method
		"secrets disable secret",                      // Disable the default secret engine
		"secrets enable -version=1 -path=secret kv",   // Enable the kv secret engine at version 1
		"write --force auth/approle/role/myrole",      // Create a role
		"write secret/testing top_secret=password123", // Create a secret
	))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
