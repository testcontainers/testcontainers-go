package compose_test

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleNewDockerComposeWith() {
	// defineComposeFile {
	composeContent := `services:
  nginx:
    image: nginx:stable-alpine
    environment:
      bar: ${bar}
      foo: ${foo}
    ports:
      - "8081:80"
  mysql:
    image: mysql:8.0.36
    environment:
      - MYSQL_DATABASE=db
      - MYSQL_ROOT_PASSWORD=my-secret-pw
    ports:
     - "3307:3306"
`
	// }

	// defineStackWithOptions {
	stack, err := compose.NewDockerComposeWith(
		compose.StackIdentifier("test"),
		compose.WithStackReaders(strings.NewReader(composeContent)),
	)
	if err != nil {
		log.Printf("Failed to create stack: %v", err)
		return
	}
	// }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// upComposeStack {
	err = stack.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		WaitForService("nginx", wait.ForListeningPort("80/tcp")).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("Failed to start stack: %v", err)
		return
	}
	defer func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %v", err)
		}
	}()
	// }

	// getServiceNames {
	serviceNames := stack.Services()
	// }

	// both services are started
	fmt.Println(len(serviceNames))
	fmt.Println(slices.Contains(serviceNames, "nginx"))
	fmt.Println(slices.Contains(serviceNames, "mysql"))

	// nginx container is started
	// getServiceContainer {
	nginxContainer, err := stack.ServiceContainer(context.Background(), "nginx")
	if err != nil {
		log.Printf("Failed to get container: %v", err)
		return
	}
	// }

	inspect, err := nginxContainer.Inspect(context.Background())
	if err != nil {
		log.Printf("Failed to inspect container: %v", err)
		return
	}

	// the nginx container has the correct environment variables
	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	for k, v := range present {
		keyVal := k + "=" + v
		fmt.Println(slices.Contains(inspect.Config.Env, keyVal))
	}

	// Output:
	// 2
	// true
	// true
	// true
	// true
}

func ExampleNewDockerComposeWith_waitForService() {
	composeContent := `services:
  nginx:
    image: nginx:stable-alpine
    environment:
      bar: ${bar}
      foo: ${foo}
    ports:
      - "8081:80"
`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stack, err := compose.NewDockerComposeWith(compose.WithStackReaders(strings.NewReader(composeContent)))
	if err != nil {
		log.Printf("Failed to create stack: %v", err)
		return
	}

	err = stack.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("Failed to start stack: %v", err)
		return
	}
	defer func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %v", err)
		}
	}()

	serviceNames := stack.Services()

	fmt.Println(serviceNames)

	// Output:
	// [nginx]
}
