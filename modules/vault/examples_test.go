package vault_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func ExampleRunContainer() {
	// runVaultContainer {
	ctx := context.Background()

	vaultContainer, err := vault.RunContainer(ctx)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withToken() {
	// runVaultContainerWithToken {
	ctx := context.Background()

	vaultContainer, err := vault.RunContainer(ctx, vault.WithToken("MyToKeN"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	cmds := []string{
		"vault", "kv", "put", "secret/test", "value=123",
	}
	exitCode, _, err := vaultContainer.Exec(ctx, cmds, exec.Multiplexed())
	if err != nil {
		panic(err)
	}

	fmt.Println(exitCode)

	// Output:
	// true
	// 0
}

func ExampleRunContainer_withInitCommand() {
	// runVaultContainerWithInitCommand {
	ctx := context.Background()

	vaultContainer, err := vault.RunContainer(ctx, vault.WithToken("MyToKeN"), vault.WithInitCommand(
		"auth enable approle",                         // Enable the approle auth method
		"secrets disable secret",                      // Disable the default secret engine
		"secrets enable -version=1 -path=secret kv",   // Enable the kv secret engine at version 1
		"write --force auth/approle/role/myrole",      // Create a role
		"write secret/testing top_secret=password123", // Create a secret
	))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := vaultContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
