package testcontainers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mysqlImage        = "mysql:8.0.36"
	nginxDelayedImage = "menedev/delayed-nginx:1.15.2"
	nginxImage        = "nginx"
	nginxAlpineImage  = "nginx:alpine"
	alpineImage       = "alpine"
	nginxDefaultPort  = "80/tcp"
	nginxHighPort     = "8080/tcp"
	golangImage       = "golang"
	daemonMaxVersion  = "1.41"
)

var providerType = ProviderDocker

func init() {
	if strings.Contains(os.Getenv("DOCKER_HOST"), "podman.sock") {
		providerType = ProviderPodman
	}
}

func TestContainerWithHostNetworkOptions(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)

	absPath, err := filepath.Abs(filepath.Join("testdata", "nginx-highport.conf"))
	require.NoError(t, err)

	opts := []ContainerCustomizer{
		WithExposedPorts(nginxHighPort),
		WithFiles(ContainerFile{
			HostFilePath:      absPath,
			ContainerFilePath: "/etc/nginx/conf.d/default.conf",
		}),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxHighPort)),
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.NetworkMode = "host"
			hc.Privileged = true
		}),
	}

	nginxC, err := Run(ctx, nginxAlpineImage, opts...)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	endpoint, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	require.NoErrorf(t, err, "Expected server endpoint")

	_, err = http.Get(endpoint)
	require.NoErrorf(t, err, "Expected OK response")
}

func TestContainerWithHostNetworkOptions_UseExposePortsFromImageConfigs(t *testing.T) {
	ctx := context.Background()

	opts := []ContainerCustomizer{
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForExposedPort()),
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Privileged = true
		}),
	}

	nginxC, err := Run(ctx, nginxImage, opts...)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	endpoint, err := nginxC.Endpoint(ctx, "http")
	require.NoErrorf(t, err, "Expected server endpoint")

	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
}

func TestContainerWithNetworkModeAndNetworkTogether(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	// skipIfDockerDesktop {
	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)
	// }

	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:    nginxImage,
			Networks: []string{"new-network"},
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.NetworkMode = "host"
			},
		},
		Started: true,
	}

	nginx, err := GenericContainer(ctx, gcr)
	CleanupContainer(t, nginx)
	if err != nil {
		// Error when NetworkMode = host and Network = []string{"bridge"}
		t.Logf("Can't use Network and NetworkMode together, %s\n", err)
	}
}

func TestContainerWithHostNetwork(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)

	absPath, err := filepath.Abs(filepath.Join("testdata", "nginx-highport.conf"))
	require.NoError(t, err)

	opts := []ContainerCustomizer{
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxHighPort)),
		WithFiles(ContainerFile{
			HostFilePath:      absPath,
			ContainerFilePath: "/etc/nginx/conf.d/default.conf",
		}),
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.NetworkMode = "host"
		}),
	}

	nginxC, err := Run(ctx, nginxAlpineImage, opts...)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	portEndpoint, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	require.NoErrorf(t, err, "Expected port endpoint %s", portEndpoint)
	t.Log(portEndpoint)

	_, err = http.Get(portEndpoint)
	require.NoErrorf(t, err, "Expected OK response")

	host, err := nginxC.Host(ctx)
	require.NoErrorf(t, err, "Expected host %s", host)

	_, err = http.Get("http://" + host + ":8080")
	require.NoErrorf(t, err, "Expected OK response")
}

func TestContainerReturnItsContainerID(t *testing.T) {
	ctx := context.Background()
	nginxA, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	require.NotEmptyf(t, nginxA.GetContainerID(), "expected a containerID but we got an empty string.")
}

// testLogConsumer is a simple implementation of LogConsumer that logs to the test output.
// It is safe to use concurrently.
type testLogConsumer struct {
	t  *testing.T
	mx sync.Mutex
}

func (l *testLogConsumer) Accept(log Log) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.t.Log(log.LogType + ": " + strings.TrimSpace(string(log.Content)))
}

func TestContainerTerminationResetsState(t *testing.T) {
	ctx := context.Background()

	nginxA, err := Run(
		ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithLogConsumerConfig(&LogConsumerConfig{
			Consumers: []LogConsumer{&testLogConsumer{t: t}},
		}),
	)
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	err = nginxA.Terminate(ctx)
	require.NoError(t, err)
	require.Empty(t, nginxA.SessionID())

	inspect, err := nginxA.Inspect(ctx)
	require.Error(t, err)
	require.Nil(t, inspect)
}

func TestContainerStateAfterTermination(t *testing.T) {
	createContainerFn := func(ctx context.Context) (Container, error) {
		return Run(ctx, nginxAlpineImage,
			WithExposedPorts(nginxDefaultPort),
			WithLogConsumerConfig(&LogConsumerConfig{
				Consumers: []LogConsumer{&testLogConsumer{t: t}},
			}),
		)
	}

	t.Run("after-termination/nil-state", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		CleanupContainer(t, nginx)
		require.NoError(t, err)

		// terminate the container before the raw state is set
		err = nginx.Terminate(ctx)
		require.NoError(t, err)

		state, err := nginx.State(ctx)
		require.Error(t, err, "expected error from container inspect.")

		require.Nil(t, state, "expected nil container inspect.")
	})

	t.Run("termination-timeout", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		require.NoError(t, err)

		err = nginx.Start(ctx)
		require.NoError(t, err, "expected no error from container start.")

		err = nginx.Terminate(ctx, StopTimeout(5*time.Microsecond))
		require.NoError(t, err)
	})

	t.Run("after-termination/nil-state-if-raw", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		CleanupContainer(t, nginx)
		require.NoError(t, err)

		state, err := nginx.State(ctx)
		require.NoError(t, err, "unexpected error from container inspect before container termination.")
		require.NotNil(t, state, "unexpected nil container inspect before container termination.")

		// terminate the container before the raw state is set
		err = nginx.Terminate(ctx)
		require.NoError(t, err)

		state, err = nginx.State(ctx)
		require.Error(t, err, "expected error from container inspect after container termination.")
		require.Nil(t, state, "unexpected nil container inspect after container termination.")
	})
}

func TestContainerTerminationRemovesDockerImage(t *testing.T) {
	t.Run("not-built-from-Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		dockerClient, err := NewDockerClientWithOpts(ctx)
		require.NoError(t, err)
		defer dockerClient.Close()

		ctr, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
		CleanupContainer(t, ctr)
		require.NoError(t, err)

		err = ctr.Terminate(ctx)
		require.NoError(t, err)

		_, err = dockerClient.ImageInspect(ctx, nginxAlpineImage)
		require.NoErrorf(t, err, "nginx image should not have been removed")
	})

	t.Run("built-from-Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		dockerClient, err := NewDockerClientWithOpts(ctx)
		require.NoError(t, err)
		defer dockerClient.Close()

		ctr, err := Run(
			ctx, "",
			WithDockerfile(FromDockerfile{
				Context: filepath.Join(".", "testdata"),
			}),
			WithExposedPorts("6379/tcp"),
			WithWaitStrategy(wait.ForLog("Ready to accept connections")),
		)
		CleanupContainer(t, ctr)
		require.NoError(t, err)
		containerID := ctr.GetContainerID()
		resp, err := dockerClient.ContainerInspect(ctx, containerID)
		require.NoError(t, err)
		imageID := resp.Config.Image

		err = ctr.Terminate(ctx)
		require.NoError(t, err)

		_, err = dockerClient.ImageInspect(ctx, imageID)
		require.Errorf(t, err, "custom built image should have been removed")
	})
}

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	nginxB, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxB)
	require.NoError(t, err)

	endpointA, err := nginxA.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpointA)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)

	endpointB, err := nginxB.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err = http.Get(endpointB)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
}

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)

	networkIP, err := nginxC.ContainerIP(ctx)
	require.NoError(t, err)
	require.NotEmptyf(t, networkIP, "Expected an IP address, got %v", networkIP)

	networkAliases, err := nginxC.NetworkAliases(ctx)
	require.NoError(t, err)
	require.Lenf(t, networkAliases, 1, "Expected number of connected networks %d. Got %d.", 0, len(networkAliases))
	require.Contains(t, networkAliases, "bridge")
	require.Emptyf(t, networkAliases["bridge"], "Expected number of aliases for 'bridge' network %d. Got %d.", 0, len(networkAliases["bridge"]))
}

func TestContainerCreationWithName(t *testing.T) {
	ctx := context.Background()

	creationName := fmt.Sprintf("%s_%d", "test_container", time.Now().Unix())
	expectedName := "/" + creationName // inspect adds '/' in the beginning

	// avoid cyclic import with the network package by defining the anonymous function here
	withBridgeNetwork := func() CustomizeRequestOption {
		return func(req *GenericContainerRequest) error {
			req.Networks = append(req.Networks, "bridge")
			return nil
		}
	}

	nginxC, err := Run(
		ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxDefaultPort)),
		WithName(creationName),
		withBridgeNetwork(),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	inspect, err := nginxC.Inspect(ctx)
	require.NoError(t, err)

	name := inspect.Name
	require.Equalf(t, expectedName, name, "Expected container name '%s'. Got '%s'.", expectedName, name)

	networks, err := nginxC.Networks(ctx)
	require.NoError(t, err)
	require.Lenf(t, networks, 1, "Expected networks 1. Got '%d'.", len(networks))
	network := networks[0]
	switch providerType {
	case ProviderDocker:
		require.Equalf(t, Bridge, network, "Expected network name '%s'. Got '%s'.", Bridge, network)
	case ProviderPodman:
		require.Equalf(t, Podman, network, "Expected network name '%s'. Got '%s'.", Podman, network)
	}

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
}

func TestContainerCreationAndWaitForListeningPortLongEnough(t *testing.T) {
	ctx := context.Background()

	// delayed-nginx will wait 2s before opening port
	nginxC, err := Run(ctx, nginxDelayedImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithPort(nginxDefaultPort)), // default startupTimeout is 60s
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)
	resp, err := http.Get(origin)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
}

func TestContainerCreationTimesOut(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := Run(ctx, nginxDelayedImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort).WithStartupTimeout(1*time.Second)),
	)
	CleanupContainer(t, nginxC)

	require.Errorf(t, err, "Expected timeout")
}

func TestContainerRespondsWithHttp200ForIndex(t *testing.T) {
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)
	resp, err := http.Get(origin)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
}

func TestContainerCreationTimesOutWithHttp(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := Run(ctx, nginxDelayedImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(time.Millisecond*500)),
	)
	CleanupContainer(t, nginxC)
	require.Error(t, err)
}

func TestContainerCreationWaitsForLogContextTimeout(t *testing.T) {
	ctx := context.Background()
	c, err := Run(ctx, mysqlImage,
		WithExposedPorts("3306/tcp", "33060/tcp"),
		WithEnv(map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		}),
		WithWaitStrategy(wait.ForLog("test context timeout").WithStartupTimeout(1*time.Second)),
	)
	CleanupContainer(t, c)
	require.Error(t, err)
}

func TestContainerCreationWaitsForLog(t *testing.T) {
	ctx := context.Background()
	mysqlC, err := Run(ctx, mysqlImage,
		WithExposedPorts("3306/tcp", "33060/tcp"),
		WithEnv(map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		}),
		WithWaitStrategy(wait.ForLog("port: 3306  MySQL Community Server - GPL")),
	)
	CleanupContainer(t, mysqlC)
	require.NoError(t, err)
}

func Test_BuildContainerFromDockerfileWithBuildArgs(t *testing.T) {
	ctx := context.Background()

	// fromDockerfileWithBuildArgs {
	ba := "build args value"
	opts := []ContainerCustomizer{
		WithExposedPorts("8080/tcp"),
		WithWaitStrategy(wait.ForLog("ready")),
		WithDockerfile(FromDockerfile{
			Context:    filepath.Join(".", "testdata"),
			Dockerfile: "args.Dockerfile",
			BuildArgs: map[string]*string{
				"FOO": &ba,
			},
		}),
	}
	c, err := Run(ctx, "", opts...)
	// }
	CleanupContainer(t, c)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	resp, err := http.Get(ep + "/env")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Equal(t, ba, string(body))
}

func Test_BuildContainerFromDockerfileWithBuildLog(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	oldStderr := os.Stderr
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = oldStderr
	})

	ctx := context.Background()

	opts := []ContainerCustomizer{
		WithDockerfile(FromDockerfile{
			Context:       filepath.Join(".", "testdata"),
			Dockerfile:    "buildlog.Dockerfile",
			PrintBuildLog: true,
		}),
	}
	c, err := Run(ctx, "", opts...)
	CleanupContainer(t, c)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	temp := strings.Split(string(out), "\n")
	require.NotEmpty(t, temp)
	require.Regexpf(t, `^Step\s*1/\d+\s*:\s*FROM alpine$`, temp[0], "Expected stdout first line to be %s. Got '%s'.", "Step 1/* : FROM alpine", temp[0])
}

func Test_BuildContainerFromDockerfileWithBuildLogWriter(t *testing.T) {
	var buffer bytes.Buffer

	ctx := context.Background()

	opts := []ContainerCustomizer{
		WithDockerfile(FromDockerfile{
			Context:        filepath.Join(".", "testdata"),
			Dockerfile:     "buildlog.Dockerfile",
			BuildLogWriter: &buffer,
		}),
	}
	c, err := Run(ctx, "", opts...)
	CleanupContainer(t, c)
	require.NoError(t, err)

	out := buffer.String()
	temp := strings.Split(out, "\n")
	require.NotEmpty(t, temp)
	require.Regexpf(t, `^Step\s*1/\d+\s*:\s*FROM alpine$`, temp[0], "Expected stdout first line to be %s. Got '%s'.", "Step 1/* : FROM alpine", temp[0])
}

func TestContainerCreationWaitsForLogAndPortContextTimeout(t *testing.T) {
	ctx := context.Background()
	c, err := Run(
		ctx, mysqlImage,
		WithExposedPorts("3306/tcp", "33060/tcp"),
		WithEnv(map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		}),
		WithWaitStrategy(wait.ForAll(
			wait.ForLog("I love testcontainers-go"),
			wait.ForListeningPort("3306/tcp"),
		)),
	)
	CleanupContainer(t, c)
	require.Errorf(t, err, "Expected timeout")
}

func TestContainerCreationWaitingForHostPort(t *testing.T) {
	ctx := context.Background()
	// exposePorts {
	nginx, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
	)
	// }
	CleanupContainer(t, nginx)
	require.NoError(t, err)
}

func TestContainerCreationWaitingForHostPortWithoutBashThrowsAnError(t *testing.T) {
	ctx := context.Background()
	nginx, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginx)
	require.NoError(t, err)
}

func TestCMD(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in a command to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	c, err := Run(ctx, alpineImage,
		WithCmd("echo", "command override!"),
		WithWaitStrategy(wait.ForAll(
			wait.ForLog("command override!"),
		)),
	)
	CleanupContainer(t, c)
	require.NoError(t, err)
}

func TestEntrypoint(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in an entrypoint to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	c, err := Run(ctx, alpineImage,
		WithEntrypoint("echo", "entrypoint override!"),
		WithWaitStrategy(wait.ForAll(
			wait.ForLog("entrypoint override!"),
		)),
	)
	CleanupContainer(t, c)
	require.NoError(t, err)
}

func TestWorkingDir(t *testing.T) {
	/*
		print the current working directory to ensure that
		we can specify working directory in the
		ContainerRequest and it will be used for the container
	*/

	ctx := context.Background()

	c, err := Run(ctx, alpineImage,
		WithEntrypoint("pwd"),
		WithWaitStrategy(wait.ForAll(
			wait.ForLog("/var/tmp/test"),
		)),
		WithConfigModifier(func(c *container.Config) {
			c.WorkingDir = "/var/tmp/test"
		}),
	)
	CleanupContainer(t, c)
	require.NoError(t, err)
}

func ExampleDockerProvider_CreateContainer() {
	ctx := context.Background()
	nginxC, err := Run(
		ctx, nginxAlpineImage,
		WithExposedPorts("80/tcp"),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create container: %s", err)
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

func ExampleContainer_Host() {
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts("80/tcp"),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return
	}
	// containerHost {
	ip, err := nginxC.Host(ctx)
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return
	}
	// }
	fmt.Println(ip)

	state, err := nginxC.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// localhost
	// true
}

func ExampleContainer_Start() {
	ctx := context.Background()
	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts("80/tcp"),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return
	}

	if err = nginxC.Start(ctx); err != nil {
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

func ExampleContainer_Stop() {
	ctx := context.Background()
	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts("80/tcp"),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create and start container: %s", err)
		return
	}

	fmt.Println("Container has been started")
	timeout := 10 * time.Second
	if err = nginxC.Stop(ctx, &timeout); err != nil {
		log.Printf("failed to terminate container: %s", err)
		return
	}

	fmt.Println("Container has been stopped")

	// Output:
	// Container has been started
	// Container has been stopped
}

func ExampleContainer_MappedPort() {
	ctx := context.Background()
	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts("80/tcp"),
		WithWaitStrategy(wait.ForHTTP("/").WithStartupTimeout(10*time.Second)),
	)
	defer func() {
		if err := TerminateContainer(nginxC); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to create and start container: %s", err)
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

func TestContainerCreationWithVolumeAndFileWritingToIt(t *testing.T) {
	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Create the volume.
	volumeName := "volumeName"

	// Create the container that writes into the mounted volume.
	bashC, err := Run(ctx, "bash:5.2.26",
		WithFiles(ContainerFile{
			HostFilePath:      absPath,
			ContainerFilePath: "/hello.sh",
			FileMode:          0o700,
		}),
		WithMounts(VolumeMount(volumeName, "/data")),
		WithCmd("bash", "/hello.sh"),
		WithWaitStrategy(wait.ForLog("done")),
	)
	CleanupContainer(t, bashC, RemoveVolumes(volumeName))
	require.NoError(t, err)
}

func TestContainerCreationWithVolumeCleaning(t *testing.T) {
	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Create the volume.
	volumeName := "volumeName"

	// Create the container that writes into the mounted volume.
	bashC, err := Run(ctx, "bash:5.2.26",
		WithFiles(ContainerFile{
			HostFilePath:      absPath,
			ContainerFilePath: "/hello.sh",
			FileMode:          0o700,
		}),
		WithMounts(VolumeMount(volumeName, "/data")),
		WithCmd("bash", "/hello.sh"),
		WithWaitStrategy(wait.ForLog("done")),
	)
	require.NoError(t, err)
	err = bashC.Terminate(ctx, RemoveVolumes(volumeName))
	CleanupContainer(t, bashC, RemoveVolumes(volumeName))
	require.NoError(t, err)
}

func TestContainerTerminationOptions(t *testing.T) {
	t.Run("volumes", func(t *testing.T) {
		var options TerminateOptions
		RemoveVolumes("vol1", "vol2")(&options)
		require.Equal(t, TerminateOptions{
			volumes: []string{"vol1", "vol2"},
		}, options)
	})
	t.Run("stop-timeout", func(t *testing.T) {
		var options TerminateOptions
		timeout := 11 * time.Second
		StopTimeout(timeout)(&options)
		require.Equal(t, TerminateOptions{
			stopTimeout: &timeout,
		}, options)
	})

	t.Run("all", func(t *testing.T) {
		var options TerminateOptions
		timeout := 9 * time.Second
		StopTimeout(timeout)(&options)
		RemoveVolumes("vol1", "vol2")(&options)
		require.Equal(t, TerminateOptions{
			stopTimeout: &timeout,
			volumes:     []string{"vol1", "vol2"},
		}, options)
	})
}

func TestContainerWithTmpFs(t *testing.T) {
	ctx := context.Background()
	ctr, err := Run(ctx, "busybox",
		WithCmd("sleep", "10"),
		WithTmpfs(map[string]string{"/testtmpfs": "rw"}),
	)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	path := "/testtmpfs/test.file"

	// exec_reader_example {
	c, reader, err := ctr.Exec(ctx, []string{"ls", path})
	require.NoError(t, err)
	require.Equalf(t, 1, c, "File %s should not have existed, expected return code 1, got %v", path, c)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	require.NoError(t, err)

	// See the logs from the command execution.
	t.Log(buf.String())
	// }

	// exec_example {
	c, _, err = ctr.Exec(ctx, []string{"touch", path})
	require.NoError(t, err)
	require.Zerof(t, c, "File %s should have been created successfully, expected return code 0, got %v", path, c)
	// }

	c, _, err = ctr.Exec(ctx, []string{"ls", path})
	require.NoError(t, err)
	require.Zerof(t, c, "File %s should exist, expected return code 0, got %v", path, c)
}

func TestContainerContextCancellation(t *testing.T) {
	t.Run("image-not-found/no-propagate-error", func(t *testing.T) {
		ctr, err := Run(context.Background(), "postgres:nonexistent-version")
		CleanupContainer(t, ctr)

		require.ErrorIs(t, err, errdefs.ErrNotFound, "the error should have been an errdefs.ErrNotFound: %v", err)
	})

	t.Run("propagate-error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		c, err := Run(ctx, "postgres:12", WithWaitStrategy(wait.ForLog("log")))
		CleanupContainer(t, c)
		require.ErrorIsf(t, err, ctx.Err(), "err should be a ctx cancelled error %v", err)
	})
}

func TestContainerCustomPlatformImage(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Incompatible Docker API version for Podman")
	}
	t.Run("non-existent-platform", func(t *testing.T) {
		t.Parallel()
		nonExistentPlatform := "windows/arm12"
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		c, err := Run(ctx, "redis:latest", WithImagePlatform(nonExistentPlatform), WithNoStart())
		CleanupContainer(t, c)
		require.Error(t, err)
	})

	t.Run("valid-platform", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, err := Run(ctx, "mysql:8.0.36", WithImagePlatform("linux/amd64"), WithNoStart())
		CleanupContainer(t, c)
		require.NoError(t, err)

		dockerCli, err := NewDockerClientWithOpts(ctx)
		require.NoError(t, err)
		defer dockerCli.Close()

		ctr, err := dockerCli.ContainerInspect(ctx, c.GetContainerID())
		require.NoError(t, err)

		img, err := dockerCli.ImageInspect(ctx, ctr.Image)
		require.NoError(t, err)
		require.Equal(t, "linux", img.Os)
		require.Equal(t, "amd64", img.Architecture)
	})
}

func TestContainerWithCustomHostname(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("some-nginx-%s-%d", t.Name(), rand.Int())
	hostname := fmt.Sprintf("my-nginx-%s-%d", t.Name(), rand.Int())

	ctr, err := Run(ctx, nginxImage,
		WithName(name),
		WithConfigModifier(func(c *container.Config) {
			c.Hostname = hostname
		}),
	)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	actualHostname := readHostname(t, ctr.GetContainerID())
	require.Equalf(t, actualHostname, hostname, "expected hostname %s, got %s", hostname, actualHostname)
}

func TestContainerInspect_RawInspectIsCleanedOnStop(t *testing.T) {
	ctr, err := Run(context.Background(), nginxImage)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	inspect, err := ctr.Inspect(context.Background())
	require.NoError(t, err)

	require.NotEmpty(t, inspect.ID)

	require.NoError(t, ctr.Stop(context.Background(), nil))
}

func readHostname(tb testing.TB, containerID string) string {
	tb.Helper()
	containerClient, err := NewDockerClientWithOpts(context.Background())
	require.NoErrorf(tb, err, "Failed to create Docker client")
	defer containerClient.Close()

	containerDetails, err := containerClient.ContainerInspect(context.Background(), containerID)
	require.NoErrorf(tb, err, "Failed to inspect container")

	return containerDetails.Config.Hostname
}

func TestDockerContainerCopyFileToContainer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		copiedFileName string
	}{
		{
			name:           "copy-file",
			copiedFileName: "/hello_copy.sh",
		},
		{
			name:           "copy-file-with-dir",
			copiedFileName: "/test/hello_copy.sh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := Run(
				ctx, nginxImage,
				WithExposedPorts(nginxDefaultPort),
				WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
			)
			CleanupContainer(t, nginxC)
			require.NoError(t, err)

			_ = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "hello.sh"), tc.copiedFileName, 700)
			c, _, err := nginxC.Exec(ctx, []string{"bash", tc.copiedFileName})
			require.NoError(t, err)
			require.Zerof(t, c, "File %s should exist, expected return code 0, got %v", tc.copiedFileName, c)
		})
	}
}

func TestDockerContainerCopyDirToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	p := filepath.Join(".", "testdata", "Dockerfile")
	err = nginxC.CopyDirToContainer(ctx, p, "/tmp/testdata/Dockerfile", 700)
	require.Error(t, err) // copying a file using the directory method will raise an error

	p = filepath.Join(".", "testdata")
	err = nginxC.CopyDirToContainer(ctx, p, "/tmp/testdata", 700)
	require.NoError(t, err)

	assertExtractedFiles(t, ctx, nginxC, p, "/tmp/testdata/")
}

func TestDockerCreateContainerWithFiles(t *testing.T) {
	ctx := context.Background()
	hostFileName := filepath.Join(".", "testdata", "hello.sh")
	copiedFileName := "/hello_copy.sh"
	tests := []struct {
		name   string
		files  []ContainerFile
		errMsg string
	}{
		{
			name: "success-copy",
			files: []ContainerFile{
				{
					HostFilePath:      hostFileName,
					ContainerFilePath: copiedFileName,
					FileMode:          0o700,
				},
			},
		},
		{
			name: "host-file-not-found",
			files: []ContainerFile{
				{
					HostFilePath:      hostFileName + "123",
					ContainerFilePath: copiedFileName,
					FileMode:          0o700,
				},
			},
			errMsg: "can't copy " +
				hostFileName + "123 to container: open " + hostFileName + "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := Run(ctx, nginxImage,
				WithExposedPorts(nginxDefaultPort),
				WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
				WithFiles(tc.files...),
				WithNoStart(),
			)
			CleanupContainer(t, nginxC)

			if err != nil {
				require.ErrorContains(t, err, tc.errMsg)
			} else {
				for _, f := range tc.files {
					require.NoError(t, err)

					hostFileData, err := os.ReadFile(f.HostFilePath)
					require.NoError(t, err)

					fd, err := nginxC.CopyFileFromContainer(ctx, f.ContainerFilePath)
					require.NoError(t, err)
					defer fd.Close()
					containerFileData, err := io.ReadAll(fd)
					require.NoError(t, err)

					require.Equal(t, hostFileData, containerFileData)
				}
			}
		})
	}
}

func TestDockerCreateContainerWithDirs(t *testing.T) {
	ctx := context.Background()
	hostDirName := "testdata"

	abs, err := filepath.Abs(filepath.Join(".", hostDirName))
	require.NoError(t, err)

	tests := []struct {
		name     string
		dir      ContainerFile
		hasError bool
	}{
		{
			name: "directory-with-absolute-path/success",
			dir: ContainerFile{
				HostFilePath:      abs,
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          0o700,
			},
			hasError: false,
		},
		{
			name: "directory/success",
			dir: ContainerFile{
				HostFilePath:      filepath.Join(".", hostDirName),
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          0o700,
			},
			hasError: false,
		},
		{
			name: "host-dir-not-found/error",
			dir: ContainerFile{
				HostFilePath:      filepath.Join(".", "testdata123"), // does not exist
				ContainerFilePath: "/tmp/" + hostDirName,             // the parent dir must exist
				FileMode:          0o700,
			},
			hasError: true,
		},
		{
			name: "container-dir-not-found/error",
			dir: ContainerFile{
				HostFilePath:      filepath.Join(".", hostDirName),
				ContainerFilePath: "/parent-does-not-exist/testdata123", // does not exist
				FileMode:          0o700,
			},
			hasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := Run(ctx, nginxImage,
				WithExposedPorts(nginxDefaultPort),
				WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
				WithFiles(tc.dir),
				WithNoStart(),
			)
			CleanupContainer(t, nginxC)

			require.Equal(t, (err != nil), tc.hasError)
			if err == nil {
				dir := tc.dir

				assertExtractedFiles(t, ctx, nginxC, dir.HostFilePath, dir.ContainerFilePath)
			}
		})
	}
}

func TestDockerContainerCopyToContainer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		copiedFileName string
	}{
		{
			name:           "success-copy",
			copiedFileName: "hello_copy.sh",
		},
		{
			name:           "success-copy-with-dir",
			copiedFileName: "/test/hello_copy.sh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := Run(ctx, nginxImage,
				WithExposedPorts(nginxDefaultPort),
				WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
			)
			CleanupContainer(t, nginxC)
			require.NoError(t, err)

			fileContent, err := os.ReadFile(filepath.Join(".", "testdata", "hello.sh"))
			require.NoError(t, err)
			err = nginxC.CopyToContainer(ctx, fileContent, tc.copiedFileName, 700)
			require.NoError(t, err)
			c, _, err := nginxC.Exec(ctx, []string{"bash", tc.copiedFileName})
			require.NoError(t, err)
			require.Zerof(t, c, "File %s should exist, expected return code 0, got %v", tc.copiedFileName, c)
		})
	}
}

func TestDockerContainerCopyFileFromContainer(t *testing.T) {
	fileContent, err := os.ReadFile(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	copiedFileName := "hello_copy.sh"
	_ = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "hello.sh"), "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	require.NoError(t, err)
	require.Zerof(t, c, "File %s should exist, expected return code 0, got %v", copiedFileName, c)

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	require.NoError(t, err)
	defer reader.Close()

	fileContentFromContainer, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, fileContent, fileContentFromContainer)
}

func TestDockerContainerCopyEmptyFileFromContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := Run(ctx, nginxImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	copiedFileName := "hello_copy.sh"
	_ = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "empty.sh"), "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	require.NoError(t, err)
	require.Zerof(t, c, "File %s should exist, expected return code 0, got %v", copiedFileName, c)

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	require.NoError(t, err)
	defer reader.Close()

	fileContentFromContainer, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Empty(t, fileContentFromContainer)
}

func TestDockerContainerResources(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Rootless Podman does not support setting rlimit")
	}
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Rootless Docker does not support setting rlimit")
	}

	ctx := context.Background()

	expected := []*container.Ulimit{
		{
			Name: "memlock",
			Hard: -1,
			Soft: -1,
		},
		{
			Name: "nofile",
			Hard: 65536,
			Soft: 65536,
		},
	}

	nginxC, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Resources = container.Resources{
				Ulimits: expected,
			}
		}),
	)
	CleanupContainer(t, nginxC)
	require.NoError(t, err)

	c, err := NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer c.Close()

	containerID := nginxC.GetContainerID()

	resp, err := c.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	require.Equal(t, expected, resp.HostConfig.Ulimits)
}

func TestContainerCapAdd(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Rootless Podman does not support setting cap-add/cap-drop")
	}

	ctx := context.Background()

	expected := "CAP_IPC_LOCK"

	nginx, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.CapAdd = []string{expected}
		}),
	)
	CleanupContainer(t, nginx)
	require.NoError(t, err)

	dockerClient, err := NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer dockerClient.Close()

	containerID := nginx.GetContainerID()
	resp, err := dockerClient.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	require.Equal(t, strslice.StrSlice{expected}, resp.HostConfig.CapAdd)
}

func TestContainerRunningCheckingStatusCode(t *testing.T) {
	ctx := context.Background()
	influx, err := Run(ctx, "influxdb:1.8.10-alpine",
		WithExposedPorts("8086/tcp"),
		WithImagePlatform("linux/amd64"), // influxdb doesn't provide an alpine+arm build (https://github.com/influxdata/influxdata-docker/issues/335)
		WithWaitStrategy(wait.ForAll(
			wait.ForHTTP("/ping").WithPort("8086/tcp").WithStatusCodeMatcher(
				func(status int) bool {
					return status == http.StatusNoContent
				},
			),
		)),
	)
	CleanupContainer(t, influx)
	require.NoError(t, err)
}

func TestContainerWithUserID(t *testing.T) {
	const expectedUserID = "60125"

	ctx := context.Background()
	ctr, err := Run(ctx, alpineImage,
		WithCmd("sh", "-c", "id -u"),
		WithWaitStrategy(wait.ForExit()),
		WithConfigModifier(func(c *container.Config) {
			c.User = expectedUserID
		}),
	)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	r, err := ctr.Logs(ctx)
	require.NoError(t, err)
	defer r.Close()
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	actual := regexp.MustCompile(`\D+`).ReplaceAllString(string(b), "")
	require.Equal(t, expectedUserID, actual)
}

func TestContainerWithNoUserID(t *testing.T) {
	ctx := context.Background()
	ctr, err := Run(ctx, alpineImage,
		WithCmd("sh", "-c", "id -u"),
		WithWaitStrategy(wait.ForExit()),
	)
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	r, err := ctr.Logs(ctx)
	require.NoError(t, err)
	defer r.Close()
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	actual := regexp.MustCompile(`\D+`).ReplaceAllString(string(b), "")
	require.Equal(t, "0", actual)
}

func TestGetGatewayIP(t *testing.T) {
	// When using docker compose with DinD mode, and using host port or http wait strategy
	// It's need to invoke GetGatewayIP for get the host
	provider, err := providerType.GetProvider(WithLogger(log.TestLogger(t)))
	require.NoError(t, err)
	defer provider.Close()

	dockerProvider, ok := provider.(*DockerProvider)
	if !ok {
		t.Skip("provider is not a DockerProvider")
	}

	ip, err := dockerProvider.GetGatewayIP(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ip)
}

func TestNetworkModeWithContainerReference(t *testing.T) {
	ctx := context.Background()
	nginxA, err := Run(ctx, nginxAlpineImage)
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	networkMode := fmt.Sprintf("container:%v", nginxA.GetContainerID())
	nginxB, err := Run(ctx, nginxAlpineImage,
		WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.NetworkMode = container.NetworkMode(networkMode)
		}),
	)
	CleanupContainer(t, nginxB)
	require.NoError(t, err)
}

// creates a temporary dir in which the files will be extracted. Then it will compare the bytes of each file in the source with the bytes from the copied-from-container file
func assertExtractedFiles(t *testing.T, ctx context.Context, container Container, hostFilePath string, containerFilePath string) {
	t.Helper()
	// create all copied files into a temporary dir
	tmpDir := t.TempDir()

	// compare the bytes of each file in the source with the bytes from the copied-from-container file
	srcFiles, err := os.ReadDir(hostFilePath)
	require.NoError(t, err)

	for _, srcFile := range srcFiles {
		if srcFile.IsDir() {
			continue
		}
		srcBytes, err := os.ReadFile(filepath.Join(hostFilePath, srcFile.Name()))
		if err != nil {
			require.NoError(t, err)
		}

		fp := filepath.Join(containerFilePath, srcFile.Name())
		// copy file by file, as there is a limitation in the Docker client to copy an entire directory from the container
		// paths for the container files are using Linux path separators
		fd, err := container.CopyFileFromContainer(ctx, fp)
		require.NoError(t, err, "Path not found in container: %s", fp)
		defer fd.Close()

		targetPath := filepath.Join(tmpDir, srcFile.Name())
		dst, err := os.Create(targetPath)
		if err != nil {
			require.NoError(t, err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, fd)
		if err != nil {
			require.NoError(t, err)
		}

		untarBytes, err := os.ReadFile(targetPath)
		if err != nil {
			require.NoError(t, err)
		}
		require.Equal(t, srcBytes, untarBytes)
	}
}

func TestDockerProviderFindContainerByName(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDockerProvider(WithLogger(log.TestLogger(t)))
	require.NoError(t, err)
	defer provider.Close()

	c1, err := Run(ctx, nginxAlpineImage,
		WithName("test"),
		WithWaitStrategy(wait.ForExposedPort()),
	)
	CleanupContainer(t, c1)
	require.NoError(t, err)

	c1Inspect, err := c1.Inspect(ctx)
	require.NoError(t, err)
	CleanupContainer(t, c1)

	c1Name := c1Inspect.Name

	c2, err := Run(ctx, nginxAlpineImage,
		WithName("test2"),
		WithWaitStrategy(wait.ForExposedPort()),
	)
	CleanupContainer(t, c2)
	require.NoError(t, err)

	c, err := provider.findContainerByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Contains(t, c.Names, c1Name)
}

func TestImageBuiltFromDockerfile_KeepBuiltImage(t *testing.T) {
	tests := []struct {
		keepBuiltImage bool
	}{
		{keepBuiltImage: true},
		{keepBuiltImage: false},
	}

	for _, tt := range tests {
		t.Run(strconv.FormatBool(tt.keepBuiltImage), func(t *testing.T) {
			ctx := context.Background()
			// Set up CLI.
			provider, err := NewDockerProvider()
			require.NoError(t, err, "get docker provider should not fail")
			defer func() { _ = provider.Close() }()
			cli := provider.Client()
			// Create container.
			c, err := Run(ctx, "",
				WithDockerfile(FromDockerfile{
					Context:    "testdata",
					Dockerfile: "echo.Dockerfile",
					KeepImage:  tt.keepBuiltImage,
				}),
			)
			CleanupContainer(t, c)
			require.NoError(t, err, "create container should not fail")
			// Get the image ID.
			containerInspect, err := c.Inspect(ctx)
			require.NoError(t, err, "container inspect should not fail")

			containerName := containerInspect.Name
			containerDetails, err := cli.ContainerInspect(ctx, containerName)
			require.NoError(t, err, "inspect container should not fail")
			containerImage := containerDetails.Image
			t.Cleanup(func() {
				_, _ = cli.ImageRemove(ctx, containerImage, image.RemoveOptions{
					Force:         true,
					PruneChildren: true,
				})
			})
			// Now, we terminate the container and check whether the image still exists.
			err = c.Terminate(ctx)
			require.NoError(t, err, "terminate container should not fail")
			_, err = cli.ImageInspect(ctx, containerImage)
			if tt.keepBuiltImage {
				require.NoError(t, err, "image should still exist")
			} else {
				require.Error(t, err, "image should not exist any more")
			}
		})
	}
}

// errMockCli is a mock implementation of client.APIClient, which is handy for simulating
// error returns in retry scenarios.
type errMockCli struct {
	client.APIClient

	err                error
	imageBuildCount    int
	containerListCount int
	imagePullCount     int
}

func (f *errMockCli) ImageBuild(_ context.Context, _ io.Reader, _ build.ImageBuildOptions) (build.ImageBuildResponse, error) {
	f.imageBuildCount++
	return build.ImageBuildResponse{Body: io.NopCloser(&bytes.Buffer{})}, f.err
}

func (f *errMockCli) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	f.containerListCount++
	return []container.Summary{{}}, f.err
}

func (f *errMockCli) ImagePull(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
	f.imagePullCount++
	return io.NopCloser(&bytes.Buffer{}), f.err
}

func (f *errMockCli) Close() error {
	return nil
}

func TestDockerProvider_BuildImage_Retries(t *testing.T) {
	tests := []struct {
		name        string
		errReturned error
		shouldRetry bool
	}{
		{
			name:        "no retry on success",
			errReturned: nil,
			shouldRetry: false,
		},
		{
			name:        "no retry when a resource is not found",
			errReturned: errdefs.ErrNotFound.WithMessage("not available"),
			shouldRetry: false,
		},
		{
			name:        "no retry when parameters are invalid",
			errReturned: errdefs.ErrInvalidArgument.WithMessage("invalid"),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access not authorized",
			errReturned: errdefs.ErrUnauthenticated.WithMessage("not authorized"),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access is forbidden",
			errReturned: errdefs.ErrPermissionDenied.WithMessage("forbidden"),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.ErrNotImplemented.WithMessage("unknown method"),
			shouldRetry: false,
		},
		{
			name:        "no retry on system error",
			errReturned: errdefs.ErrInternal.WithMessage("system error"),
			shouldRetry: false,
		},
		{
			name:        "retry on non-permanent error",
			errReturned: errors.New("whoops"),
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewDockerProvider()
			require.NoError(t, err)
			m := &errMockCli{err: tt.errReturned}
			p.client = m

			// give a chance to retry
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, err = p.BuildImage(ctx, &ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context: filepath.Join(".", "testdata", "retry"),
				},
			})
			if tt.errReturned != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Positive(t, m.imageBuildCount)
			require.Equal(t, tt.shouldRetry, m.imageBuildCount > 1)
		})
	}
}

func TestDockerProvider_waitContainerCreation_retries(t *testing.T) {
	tests := []struct {
		name        string
		errReturned error
		shouldRetry bool
	}{
		{
			name:        "no retry on success",
			errReturned: nil,
			shouldRetry: false,
		},
		{
			name:        "no retry when parameters are invalid",
			errReturned: errdefs.ErrInvalidArgument.WithMessage("invalid"),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.ErrNotImplemented.WithMessage("unknown method"),
			shouldRetry: false,
		},
		{
			name:        "retry when not found",
			errReturned: errdefs.ErrNotFound.WithMessage("not there yet"),
			shouldRetry: true,
		},
		{
			name:        "retry on non-permanent error",
			errReturned: errors.New("whoops"),
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewDockerProvider()
			require.NoError(t, err)
			m := &errMockCli{err: tt.errReturned}
			p.client = m

			// give a chance to retry
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, _ = p.waitContainerCreation(ctx, "someID")

			require.Positive(t, m.containerListCount)
			require.Equal(t, tt.shouldRetry, m.containerListCount > 1)
		})
	}
}

func TestDockerProvider_attemptToPullImage_retries(t *testing.T) {
	tests := []struct {
		name        string
		errReturned error
		shouldRetry bool
	}{
		{
			name:        "no retry on success",
			errReturned: nil,
			shouldRetry: false,
		},
		{
			name:        "no retry when a resource is not found",
			errReturned: errdefs.ErrNotFound.WithMessage("not available"),
			shouldRetry: false,
		},
		{
			name:        "no retry when parameters are invalid",
			errReturned: errdefs.ErrInvalidArgument.WithMessage("invalid"),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access not authorized",
			errReturned: errdefs.ErrUnauthenticated.WithMessage("not authorized"),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access is forbidden",
			errReturned: errdefs.ErrPermissionDenied.WithMessage("forbidden"),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.ErrNotImplemented.WithMessage("unknown method"),
			shouldRetry: false,
		},
		{
			name:        "retry on non-permanent error",
			errReturned: errors.New("whoops"),
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewDockerProvider()
			require.NoError(t, err)
			m := &errMockCli{err: tt.errReturned}
			p.client = m

			// give a chance to retry
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_ = p.attemptToPullImage(ctx, "someTag", image.PullOptions{})

			require.Positive(t, m.imagePullCount)
			require.Equal(t, tt.shouldRetry, m.imagePullCount > 1)
		})
	}
}

func TestCustomPrefixTrailingSlashIsProperlyRemovedIfPresent(t *testing.T) {
	hubPrefixWithTrailingSlash := "public.ecr.aws/"
	dockerImage := "amazonlinux/amazonlinux:2023"

	ctx := context.Background()
	c, err := Run(ctx, dockerImage, WithImageSubstitutors(newPrependHubRegistry(hubPrefixWithTrailingSlash)))
	CleanupContainer(t, c)
	require.NoError(t, err)

	require.Equal(t, fmt.Sprintf("%s%s", hubPrefixWithTrailingSlash, dockerImage), c.Image)
}

// TODO: remove this skip check when context rework is merged alongside [core.DockerEnvFile] removal.
func Test_Provider_DaemonHost_Issue2897(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, provider.Close())
	})

	orig := core.DockerEnvFile
	core.DockerEnvFile = filepath.Join(t.TempDir(), ".dockerenv")
	t.Cleanup(func() {
		core.DockerEnvFile = orig
	})

	f, err := os.Create(core.DockerEnvFile)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() {
		require.NoError(t, os.Remove(f.Name()))
	})

	errCh := make(chan error, 1)
	go func() {
		_, err := provider.DaemonHost(ctx)
		errCh <- err
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for DaemonHost")
	case err := <-errCh:
		require.NoError(t, err)
	}
}
