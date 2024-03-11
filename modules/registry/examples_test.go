package registry_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRunContainer() {
	// runRegistryContainer {
	ctx := context.Background()

	registryContainer, err := registry.RunContainer(ctx, testcontainers.WithImage("registry:2.8.3"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := registryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := registryContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withAuthentication() {
	ctx := context.Background()

	// htpasswdFile {
	registryContainer, err := registry.RunContainer(
		ctx, testcontainers.WithImage("registry:2.8.3"),
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "data")),
	)
	// }
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := registryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	registryPort, err := registryContainer.MappedPort(ctx, "5000/tcp")
	if err != nil {
		log.Fatalf("failed to get mapped port: %s", err) // nolint:gocritic
	}
	strPort := registryPort.Port()

	previousAuthConfig := os.Getenv("DOCKER_AUTH_CONFIG")

	// make sure the Docker Auth credentials are set
	// using the same as in the Docker Registry
	// testuser:testpassword
	os.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
			"localhost:`+strPort+`": { "username": "testuser", "password": "testpassword", "auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" }
		},
		"credsStore": "desktop"
	}`)
	defer func() {
		// reset the original state after the example.
		os.Unsetenv("DOCKER_AUTH_CONFIG")
		os.Setenv("DOCKER_AUTH_CONFIG", previousAuthConfig)
	}()

	// build a custom redis image from the private registry
	// it will use localhost:$exposedPort as the registry

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_PORT": &strPort,
				},
				PrintBuildLog: true,
			},
			AlwaysPullImage: true, // make sure the authentication takes place
			ExposedPorts:    []string{"6379/tcp"},
			WaitingFor:      wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to start container: %s", err) // nolint:gocritic
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	state, err := redisC.State(ctx)
	if err != nil {
		log.Fatalf("failed to get redis container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
