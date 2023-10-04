package k6_test

import (
	"context"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modules/k6"
)

func ExampleRunContainer() {
	// runK6Container {
	ctx := context.Background()

	absPath, err := filepath.Abs("./scripts/test.js")
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
	// }
}
