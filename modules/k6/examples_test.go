package k6_test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k6"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRunContainer() {
	// runK6Container {
	ctx := context.Background()

	gcr := testcontainers.GenericContainerRequest{
		ProviderType: testcontainers.ProviderDocker,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "kennethreitz/httpbin",
			ExposedPorts: []string{
				"80",
			},
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	}
	httpbin, err := testcontainers.GenericContainer(ctx, gcr)
	if err != nil {
		panic(fmt.Errorf("failed to create httpbin container %w", err))
	}

	defer func() {
		if err := httpbin.Terminate(ctx); err != nil {
			panic(fmt.Errorf("failed to terminate container: %w", err))
		}
	}()

	httpbinIP, err := httpbin.ContainerIP(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get httpbin IP: %w", err))
	}

	absPath, err := filepath.Abs(filepath.Join("scripts", "httpbin.js"))
	if err != nil {
		panic(fmt.Errorf("failed to get path to test script: %w", err))
	}

	k6, err := k6.RunContainer(
		ctx,
		k6.WithTestScript(absPath),
		k6.SetEnvVar("HTTPBIN", httpbinIP),
	)
	if err != nil {
		panic(fmt.Errorf("failed to start k6 container: %w", err))
	}

	defer func() {
		if err := k6.Terminate(ctx); err != nil {
			panic(fmt.Errorf("failed to terminate container: %w", err))
		}
	}()

	// assert the result of the test
	state, err := k6.State(ctx)
	if err != nil {
		panic(err)
	}
	if state.ExitCode != 0 {
		panic("k6 test failed")
	}
	//}
}
