package testcontainers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"

	tcimage "github.com/testcontainers/testcontainers-go/image"
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
	nginxC, err := Run(ctx, req)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_buildFromDockerfile() {
	ctx := context.Background()

	// buildFromDockerfileWithModifier {
	c, err := Run(ctx, Request{
		FromDockerfile: FromDockerfile{
			Context:    "testdata",
			Dockerfile: "target.Dockerfile",
			KeepImage:  false,
			BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
				buildOptions.Target = "target2"
			},
		},
		Started: true,
	})
	// }
	defer func() {
		if err := TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	r, err := c.Logs(ctx)
	if err != nil {
		log.Printf("failed to get logs: %v", err)
		return
	}

	logs, err := io.ReadAll(r)
	if err != nil {
		log.Printf("failed to read logs: %v", err)
		return
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
	nginxC, err := Run(ctx, req)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	// containerHost {
	ip, err := nginxC.Host(ctx)
	if err != nil {
		log.Printf("failed to get container host: %s", err)
		return
	}
	// }
	println(ip)

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
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
	nginxC, err := Run(ctx, req)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return
	}

	err = nginxC.Start(ctx)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
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
	nginxC, err := Run(ctx, req)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	fmt.Println("Container has been started")
	timeout := 10 * time.Second
	err = nginxC.Stop(ctx, &timeout)
	if err != nil {
		log.Printf("failed to stop container: %s", err)
		return
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
	nginxC, err := Run(ctx, req)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	// buildingAddresses {
	ip, _ := nginxC.Host(ctx)
	port, _ := nginxC.MappedPort(ctx, "80")
	_, _ = http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	// }

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleNew_withNetwork() {
	ctx := context.Background()

	net, err := NewNetwork(ctx,
		network.WithAttachable(),
		network.WithDriver("bridge"),
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

	nginxC, err := Run(ctx, Request{
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
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

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
	ctr, err := Run(ctx, Request{
		Image:             "alpine:latest",
		ImageSubstitutors: []tcimage.Substitutor{tcimage.DockerSubstitutor{}},
		Started:           true,
	})
	defer func() {
		if err := TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	// }
	if err != nil {
		log.Printf("could not start container: %v", err)
		return
	}

	fmt.Println(ctr.Image)

	// Output: docker.io/alpine:latest
}
