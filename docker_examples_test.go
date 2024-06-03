package testcontainers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleNew() {
	ctx := context.Background()
	req := Request{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:      true,
	}
	nginxC, _ := New(ctx, req)
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_buildFromDockerfile() {
	ctx := context.Background()

	// buildFromDockerfileWithModifier {
	c, err := New(ctx, Request{
		FromDockerfile: FromDockerfile{
			Context:       "testdata",
			Dockerfile:    "target.Dockerfile",
			PrintBuildLog: true,
			KeepImage:     false,
			BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
				buildOptions.Target = "target2"
			},
		},
		Started: true,
	})
	// }
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	r, err := c.Logs(ctx)
	if err != nil {
		log.Fatalf("failed to get logs: %v", err)
	}

	logs, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("failed to read logs: %v", err)
	}

	fmt.Println(string(logs))

	// Output: target2
}

func ExampleNew_containerHost() {
	ctx := context.Background()
	req := Request{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:      true,
	}
	nginxC, _ := New(ctx, req)
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// containerHost {
	ip, _ := nginxC.Host(ctx)
	// }
	println(ip)

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_containerStart() {
	ctx := context.Background()
	req := Request{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
	}
	nginxC, _ := New(ctx, req)
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	_ = nginxC.Start(ctx)

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_containerStop() {
	ctx := context.Background()
	req := Request{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:      true,
	}
	nginxC, _ := New(ctx, req)
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	fmt.Println("Container has been started")
	timeout := 10 * time.Second
	err := nginxC.Stop(ctx, &timeout)
	if err != nil {
		log.Fatalf("failed to stop container: %s", err) // nolint:gocritic
	}

	fmt.Println("Container has been stopped")

	// Output:
	// Container has been started
	// Container has been stopped
}

func ExampleNew_mappedPort() {
	ctx := context.Background()
	req := Request{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:      true,
	}
	nginxC, _ := New(ctx, req)
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// buildingAddresses {
	ip, _ := nginxC.Host(ctx)
	port, _ := nginxC.MappedPort(ctx, "80")
	_, _ = http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	// }

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_withNetwork() {
	ctx := context.Background()

	net, err := NewNetwork(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		// Makes the network internal only, meaning the host machine cannot access it.
		// Remove or use `network.WithDriver("bridge")` to change the network's mode.
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			log.Fatalf("failed to remove network: %s", err)
		}
	}()

	networkName := net.Name
	// }

	nginxC, _ := New(ctx, Request{
		Image: "nginx:alpine",
		ExposedPorts: []string{
			"80/tcp",
		},
		Networks: []string{
			networkName,
		},
		Started: true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	state, err := nginxC.State(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_withSubstitutors() {
	ctx := context.Background()

	// applyImageSubstitutors {
	container, err := New(ctx, Request{
		Image:             "alpine:latest",
		ImageSubstitutors: []image.Substitutor{image.DockerSubstitutor{}},
		Started:           true,
	})
	// }
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			log.Fatalf("could not terminate container: %v", err)
		}
	}()

	fmt.Println(container.Image)

	// Output: docker.io/alpine:latest
}
