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
	defer func() {
		if err := testcontainers.TerminateContainer(registryContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := registryContainer.State(context.Background())
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
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
	defer func() {
		if err := testcontainers.TerminateContainer(registryContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	registryHost, err := registryContainer.HostAddress(ctx)
	if err != nil {
		log.Printf("failed to get host: %s", err)
		return
	}

	cleanup, err := registry.SetDockerAuthConfig(registryHost, "testuser", "testpassword")
	if err != nil {
		log.Printf("failed to set docker auth config: %s", err)
		return
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
			},
			AlwaysPullImage: true, // make sure the authentication takes place
			ExposedPorts:    []string{"6379/tcp"},
			WaitingFor:      wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(redisC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := redisC.State(context.Background())
	if err != nil {
		log.Printf("failed to get redis container state: %s", err)
		return
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
	defer func() {
		if err := testcontainers.TerminateContainer(registryContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	registryHost, err := registryContainer.HostAddress(ctx)
	if err != nil {
		log.Printf("failed to get host: %s", err)
		return
	}

	// Besides, we are also setting the authentication
	// for both the registry and localhost to make sure
	// the image is pushed to the private registry.
	cleanup, err := registry.SetDockerAuthConfig(
		registryHost, "testuser", "testpassword",
		registryContainer.RegistryName, "testuser", "testpassword",
	)
	if err != nil {
		log.Printf("failed to set docker auth config: %s", err)
		return
	}
	defer cleanup()

	// build a custom redis image from the private registry,
	// using RegistryName of the container as the registry.
	// We are going to build the image with a fixed tag
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
				Repo: repo,
				Tag:  tag,
			},
			AlwaysPullImage: true, // make sure the authentication takes place
			ExposedPorts:    []string{"6379/tcp"},
			WaitingFor:      wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(redisC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	newImage := fmt.Sprintf("%s:%s", repo, tag)
	err = registryContainer.PushImage(context.Background(), fmt.Sprintf("%s:%s", repo, tag))
	if err != nil {
		log.Printf("failed to push image: %s", err)
		return
	}

	// pull a redis image from an public registry,
	// tag it specifying the local registry name,
	// and push to the private registry.

	defaultRegistryURI := "docker.io/library"
	defaultImage := "redis"
	defaultTag := "5.0-alpine"

	imageRef := fmt.Sprintf("%s/%s:%s", defaultRegistryURI, defaultImage, defaultTag)
	err = registryContainer.PullImage(ctx, imageRef)
	if err != nil {
		log.Printf("failed to pull image: %s", err)
		return
	}

	taggedImage := fmt.Sprintf("%s/%s:%s", registryContainer.RegistryName, defaultImage, defaultTag)
	err = registryContainer.TagImage(ctx, imageRef, taggedImage)
	if err != nil {
		log.Printf("failed to tag image: %s", err)
		return
	}

	err = registryContainer.PushImage(context.Background(), taggedImage)
	if err != nil {
		log.Printf("failed to push image: %s", err)
		return
	}

	// now run a container from the new image
	// But first remove the local image to avoid using the local one.

	err = registryContainer.DeleteImage(context.Background(), newImage)
	if err != nil {
		log.Printf("failed to delete image: %s", err)
		return
	}
	err = registryContainer.DeleteImage(context.Background(), taggedImage)
	if err != nil {
		log.Printf("failed to delete image: %s", err)
		return
	}

	newRedisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        taggedImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(newRedisC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container from %s: %s", taggedImage, err)
		return
	}

	state, err := newRedisC.State(context.Background())
	if err != nil {
		log.Printf("failed to get redis container state from %s: %s", taggedImage, err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
