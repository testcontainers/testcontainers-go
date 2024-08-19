package registry_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	// runRegistryContainer {
	registryContainer, err := registry.Run(context.Background(), "registry:2.8.3")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := registryContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := registryContainer.State(context.Background())
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withAuthentication() {
	// htpasswdFile {
	ctx := context.Background()
	registryContainer, err := registry.Run(
		ctx,
		"registry:2.8.3",
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

	registryHost, err := registryContainer.HostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %s", err) // nolint:gocritic
	}

	cleanup, err := registry.SetDockerAuthConfig(registryHost, "testuser", "testpassword")
	if err != nil {
		log.Fatalf("failed to set docker auth config: %s", err) // nolint:gocritic
	}
	defer cleanup()

	// build a custom redis image from the private registry,
	// using RegistryName of the container as the registry.

	redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_HOST": &registryHost,
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
		if err := redisC.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	state, err := redisC.State(context.Background())
	if err != nil {
		log.Fatalf("failed to get redis container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_pushImage() {
	ctx := context.Background()
	registryContainer, err := registry.Run(
		ctx,
		registry.DefaultImage,
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "data")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := registryContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	registryHost, err := registryContainer.HostAddress(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %s", err) // nolint:gocritic
	}

	// Besides, we are also setting the authentication
	// for both the registry and localhost to make sure
	// the image is pushed to the private registry.
	cleanup, err := registry.SetDockerAuthConfig(
		registryHost, "testuser", "testpassword",
		registryContainer.RegistryName, "testuser", "testpassword",
	)
	if err != nil {
		log.Fatalf("failed to set docker auth config: %s", err) // nolint:gocritic
	}
	defer cleanup()

	// build a custom redis image from the private registry,
	// using RegistryName of the container as the registry.
	// We are agoing to build the image with a fixed tag
	// that matches the private registry, and we are going to
	// push it again to the registry after the build.

	repo := registryContainer.RegistryName + "/customredis"
	tag := "v1.2.3"

	redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_HOST": &registryHost,
				},
				Repo:          repo,
				Tag:           tag,
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
		if err := redisC.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// pushingImage {
	// repo is localhost:32878/customredis
	// tag is v1.2.3
	err = registryContainer.PushImage(context.Background(), fmt.Sprintf("%s:%s", repo, tag))
	if err != nil {
		log.Fatalf("failed to push image: %s", err) // nolint:gocritic
	}
	// }

	newImage := fmt.Sprintf("%s:%s", repo, tag)

	// now run a container from the new image
	// But first remove the local image to avoid using the local one.

	// deletingImage {
	// newImage is customredis:v1.2.3
	err = registryContainer.DeleteImage(context.Background(), newImage)
	if err != nil {
		log.Fatalf("failed to delete image: %s", err) // nolint:gocritic
	}
	// }

	newRedisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        newImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to start container from %s: %s", newImage, err) // nolint:gocritic
	}
	defer func() {
		if err := newRedisC.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	state, err := newRedisC.State(context.Background())
	if err != nil {
		log.Fatalf("failed to get redis container state from %s: %s", newImage, err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
