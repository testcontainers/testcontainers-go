package vault_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func ExampleRun() {
	// runVaultContainer {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0")
	defer func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withToken() {
	// runVaultContainerWithToken {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0", vault.WithToken("MyToKeN"))
	defer func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	cmds := []string{
		"vault", "kv", "put", "secret/test", "value=123",
	}
	exitCode, _, err := vaultContainer.Exec(ctx, cmds, exec.Multiplexed())
	if err != nil {
		log.Printf("failed to execute command: %s", err)
		return
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
	defer func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
