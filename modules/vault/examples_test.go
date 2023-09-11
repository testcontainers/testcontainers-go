package vault_test

import (
	"context"
	"fmt"

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
