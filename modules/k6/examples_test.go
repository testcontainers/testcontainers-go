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
	// runHTTPBin {
	ctx := context.Background()

	// create a container with the httpbin application that will be the target
	// for the test script that runs in the k6 container
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
	// }

	// getHTTPBinIP {
	httpbinIP, err := httpbin.ContainerIP(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get httpbin IP: %w", err))
	}
	// }

	absPath, err := filepath.Abs(filepath.Join("scripts", "httpbin.js"))
	if err != nil {
		panic(fmt.Errorf("failed to get path to test script: %w", err))
	}

	// runK6Container {
	// run the httpbin.js test scripts passing the IP address the httpbin container
	k6, err := k6.RunContainer(
		ctx,
		k6.WithCache(),
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
	//}

	// assert the result of the test
	state, err := k6.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.ExitCode)
	// Output: 0
}
