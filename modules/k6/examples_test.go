package k6_test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modules/k6"
)

func ExampleRunContainer() {
	// runK6Container {
	ctx := context.Background()

	absPath, err := filepath.Abs("./scripts/pass.js")
	if err != nil {
		panic(err)
	}

	container, err := k6.RunContainer(ctx, k6.WithTestScript(absPath))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	// assert the result of the test
	state, err := container.State(ctx)
	if err != nil {
		panic(err)
	}
	if state.ExitCode != 0 {
		panic(fmt.Errorf("test failed with exit code %d", state.ExitCode))
	}
	// }
}
