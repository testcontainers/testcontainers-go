package testcontainers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	// Import mysql into the scope of this package (required)
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-units"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/docker/docker/errdefs"

	"github.com/docker/docker/api/types/volume"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mysqlImage       = "docker.io/mysql:8.0.30"
	nginxImage       = "docker.io/nginx"
	nginxAlpineImage = "docker.io/nginx:alpine"
	nginxDefaultPort = "80/tcp"
	nginxHighPort    = "8080/tcp"
	daemonMaxVersion = "1.41"
)

var providerType = ProviderDocker

func init() {
	if strings.Contains(os.Getenv("DOCKER_HOST"), "podman.sock") {
		providerType = ProviderPodman
	}
}

// testNetworkAliases {
func TestContainerAttachedToNewNetwork(t *testing.T) {
	aliases := []string{"alias1", "alias2", "alias3"}
	networkName := "new-network"
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			Networks: []string{
				networkName,
			},
			NetworkAliases: map[string][]string{
				networkName: aliases,
			},
		},
		Started: true,
	}

	newNetwork, err := GenericNetwork(ctx, GenericNetworkRequest{
		ProviderType: providerType,
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	nginx, err := GenericContainer(ctx, gcr)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginx)

	networks, err := nginx.Networks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networks) != 1 {
		t.Errorf("Expected networks 1. Got '%d'.", len(networks))
	}
	network := networks[0]
	if network != networkName {
		t.Errorf("Expected network name '%s'. Got '%s'.", networkName, network)
	}

	networkAliases, err := nginx.NetworkAliases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkAliases) != 1 {
		t.Errorf("Expected network aliases for 1 network. Got '%d'.", len(networkAliases))
	}

	networkAlias := networkAliases[networkName]

	require.NotEmpty(t, networkAlias)

	for _, alias := range aliases {
		require.Contains(t, networkAlias, alias)
	}

	networkIP, err := nginx.ContainerIP(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkIP) == 0 {
		t.Errorf("Expected an IP address, got %v", networkIP)
	}
}

// }

func TestContainerWithHostNetworkOptions(t *testing.T) {
	absPath, err := filepath.Abs("./testresources/nginx-highport.conf")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxAlpineImage,
			Privileged:  true,
			SkipReaper:  true,
			NetworkMode: "host",
			Mounts:      Mounts(BindMount(absPath, "/etc/nginx/conf.d/default.conf")),
			ExposedPorts: []string{
				nginxHighPort,
			},
			WaitingFor: wait.ForListeningPort(nginxHighPort),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	// host, err := nginxC.Host(ctx)
	// if err != nil {
	//	t.Errorf("Expected host %s. Got '%d'.", host, err)
	// }
	//
	endpoint, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	if err != nil {
		t.Errorf("Expected server endpoint. Got '%v'.", err)
	}

	_, err = http.Get(endpoint)
	if err != nil {
		t.Errorf("Expected OK response. Got '%d'.", err)
	}
}

func TestContainerWithHostNetworkOptions_UseExposePortsFromImageConfigs(t *testing.T) {
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:      "nginx",
			Privileged: true,
			SkipReaper: true,
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	terminateContainerOnEnd(t, ctx, nginxC)

	endpoint, err := nginxC.Endpoint(ctx, "http")
	if err != nil {
		t.Errorf("Expected server endpoint. Got '%v'.", err)
	}

	_, err = http.Get(endpoint)
	if err != nil {
		t.Errorf("Expected OK response. Got '%d'.", err)
	}
}

func TestContainerWithNetworkModeAndNetworkTogether(t *testing.T) {
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxImage,
			SkipReaper:  true,
			NetworkMode: "host",
			Networks:    []string{"new-network"},
		},
		Started: true,
	}

	nginx, err := GenericContainer(ctx, gcr)
	if err != nil {
		// Error when NetworkMode = host and Network = []string{"bridge"}
		t.Logf("Can't use Network and NetworkMode together, %s", err)
	}
	terminateContainerOnEnd(t, ctx, nginx)
}

func TestContainerWithHostNetworkOptionsAndWaitStrategy(t *testing.T) {
	ctx := context.Background()

	absPath, err := filepath.Abs("./testresources/nginx-highport.conf")
	if err != nil {
		t.Fatal(err)
	}

	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxAlpineImage,
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort(nginxHighPort),
			Mounts:      Mounts(BindMount(absPath, "/etc/nginx/conf.d/default.conf")),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	host, err := nginxC.Host(ctx)
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", host, err)
	}

	_, err = http.Get("http://" + host + ":8080")
	if err != nil {
		t.Errorf("Expected OK response. Got '%v'.", err)
	}
}

func TestContainerWithHostNetworkAndEndpoint(t *testing.T) {
	ctx := context.Background()

	absPath, err := filepath.Abs("./testresources/nginx-highport.conf")
	if err != nil {
		t.Fatal(err)
	}

	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxAlpineImage,
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort(nginxHighPort),
			Mounts:      Mounts(BindMount(absPath, "/etc/nginx/conf.d/default.conf")),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	hostN, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", hostN, err)
	}
	t.Log(hostN)

	_, err = http.Get(hostN)
	if err != nil {
		t.Errorf("Expected OK response. Got '%v'.", err)
	}
}

func TestContainerWithHostNetworkAndPortEndpoint(t *testing.T) {
	ctx := context.Background()

	absPath, err := filepath.Abs("./testresources/nginx-highport.conf")
	if err != nil {
		t.Fatal(err)
	}

	gcr := GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxAlpineImage,
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort(nginxHighPort),
			Mounts:      Mounts(BindMount(absPath, "/etc/nginx/conf.d/default.conf")),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	origin, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", origin, err)
	}
	t.Log(origin)

	_, err = http.Get(origin)
	if err != nil {
		t.Errorf("Expected OK response. Got '%d'.", err)
	}
}

func TestContainerReturnItsContainerID(t *testing.T) {
	ctx := context.Background()
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxA)

	if nginxA.GetContainerID() == "" {
		t.Errorf("expected a containerID but we got an empty string.")
	}
}

func TestContainerStartsWithoutTheReaper(t *testing.T) {
	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	var container Container
	container, err = GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			SkipReaper: true,
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	resp, err := client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", TestcontainerLabelSessionID, container.SessionID()))),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 0 {
		t.Fatal("expected zero reaper running.")
	}
}

func TestContainerStartsWithTheReaper(t *testing.T) {
	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	_, err = GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	filtersJSON := fmt.Sprintf(`{"label":{"%s":true}}`, TestcontainerLabelIsReaper)
	f, err := filters.FromJSON(filtersJSON)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.ContainerList(ctx, types.ContainerListOptions{
		Filters: f,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one reaper to be running.")
	}
}

func TestContainerTerminationResetsState(t *testing.T) {
	ctx := context.Background()

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			SkipReaper: true,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if nginxA.SessionID() != "00000000-0000-0000-0000-000000000000" {
		t.Fatal("Internal state must be reset.")
	}
	ports, err := nginxA.Ports(ctx)
	if err == nil || ports != nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerStateAfterTermination(t *testing.T) {
	createContainerFn := func(ctx context.Context) (Container, error) {
		return GenericContainer(ctx, GenericContainerRequest{
			ProviderType: providerType,
			ContainerRequest: ContainerRequest{
				Image: nginxAlpineImage,
				ExposedPorts: []string{
					nginxDefaultPort,
				},
			},
			Started: true,
		})
	}

	t.Run("Nil State after termination", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// terminate the container before the raw state is set
		err = nginx.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}

		state, err := nginx.State(ctx)
		assert.Error(t, err, "expected error from container inspect.")

		assert.Nil(t, state, "expected nil container inspect.")
	})

	t.Run("Non-nil State after termination if raw as already set", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		if err != nil {
			t.Fatal(err)
		}

		state, err := nginx.State(ctx)
		assert.NoError(t, err, "unexpected error from container inspect before container termination.")

		assert.NotNil(t, state, "unexpected nil container inspect before container termination.")

		// terminate the container before the raw state is set
		err = nginx.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}

		state, err = nginx.State(ctx)
		assert.Error(t, err, "expected error from container inspect after container termination.")

		assert.NotNil(t, state, "unexpected nil container inspect after container termination.")
	})
}

func TestContainerStopWithReaper(t *testing.T) {
	ctx := context.Background()

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxA)

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	stopTimeout := 10 * time.Second
	err = nginxA.Stop(ctx, &stopTimeout)
	if err != nil {
		t.Fatal(err)
	}

	state, err = nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != false {
		t.Fatal("The container shoud not be running")
	}
	if state.Status != "exited" {
		t.Fatal("The container shoud be in exited state")
	}
}

func TestContainerTerminationWithReaper(t *testing.T) {
	ctx := context.Background()

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = nginxA.State(ctx)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerTerminationWithoutReaper(t *testing.T) {
	ctx := context.Background()

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			SkipReaper: true,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nginxA.State(ctx)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerTerminationRemovesDockerImage(t *testing.T) {
	t.Run("if not built from Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		client, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			t.Fatal(err)
		}
		client.NegotiateAPIVersion(ctx)
		container, err := GenericContainer(ctx, GenericContainerRequest{
			ProviderType: providerType,
			ContainerRequest: ContainerRequest{
				Image: nginxAlpineImage,
				ExposedPorts: []string{
					nginxDefaultPort,
				},
				SkipReaper: true,
			},
			Started: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = container.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = client.ImageInspectWithRaw(ctx, nginxAlpineImage)
		if err != nil {
			t.Fatal("nginx image should not have been removed")
		}
	})

	t.Run("if built from Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		client, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			t.Fatal(err)
		}
		client.NegotiateAPIVersion(ctx)
		req := ContainerRequest{
			FromDockerfile: FromDockerfile{
				Context: "./testresources",
			},
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		}
		container, err := GenericContainer(ctx, GenericContainerRequest{
			ProviderType:     providerType,
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			t.Fatal(err)
		}
		containerID := container.GetContainerID()
		resp, err := client.ContainerInspect(ctx, containerID)
		if err != nil {
			t.Fatal(err)
		}
		imageID := resp.Config.Image

		err = container.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = client.ImageInspectWithRaw(ctx, imageID)
		if err == nil {
			t.Fatal("custom built image should have been removed", err)
		}
	})
}

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxA)

	nginxB, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxB)

	endpointA, err := nginxA.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpointA)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	endpointB, err := nginxB.PortEndpoint(ctx, nginxDefaultPort, "http")
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Get(endpointB)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
	networkIP, err := nginxC.ContainerIP(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkIP) == 0 {
		t.Errorf("Expected an IP address, got %v", networkIP)
	}
	networkAliases, err := nginxC.NetworkAliases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkAliases) != 1 {
		fmt.Printf("%v", networkAliases)
		t.Errorf("Expected number of connected networks %d. Got %d.", 0, len(networkAliases))
	}
	if len(networkAliases["bridge"]) != 0 {
		t.Errorf("Expected number of aliases for 'bridge' network %d. Got %d.", 0, len(networkAliases["bridge"]))
	}
}

func TestContainerIPs(t *testing.T) {
	ctx := context.Background()

	networkName := "new-network"
	newNetwork, err := GenericNetwork(ctx, GenericNetworkRequest{
		ProviderType: providerType,
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			Networks: []string{
				"bridge",
				networkName,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	ips, err := nginxC.ContainerIPs(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(ips) != 2 {
		t.Errorf("Expected two IP addresses, got %v", len(ips))
	}
}

func TestContainerCreationWithName(t *testing.T) {
	ctx := context.Background()

	creationName := fmt.Sprintf("%s_%d", "test_container", time.Now().Unix())
	expectedName := "/" + creationName // inspect adds '/' in the beginning

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort),
			Name:       creationName,
			Networks:   []string{"bridge"},
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	name, err := nginxC.Name(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if name != expectedName {
		t.Errorf("Expected container name '%s'. Got '%s'.", expectedName, name)
	}
	networks, err := nginxC.Networks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networks) != 1 {
		t.Errorf("Expected networks 1. Got '%d'.", len(networks))
	}
	network := networks[0]
	switch providerType {
	case ProviderDocker:
		if network != Bridge {
			t.Errorf("Expected network name '%s'. Got '%s'.", Bridge, network)
		}
	case ProviderPodman:
		if network != Podman {
			t.Errorf("Expected network name '%s'. Got '%s'.", Podman, network)
		}
	}

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationAndWaitForListeningPortLongEnough(t *testing.T) {
	ctx := context.Background()

	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: "docker.io/menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort), // default startupTimeout is 60s
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(origin)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationTimesOut(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: "docker.io/menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort).WithStartupTimeout(1 * time.Second),
		},
		Started: true,
	})
	if err == nil {
		t.Error("Expected timeout")
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestContainerRespondsWithHttp200ForIndex(t *testing.T) {
	ctx := context.Background()

	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForHTTP("/"),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(origin)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationTimesOutWithHttp(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: "docker.io/menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			WaitingFor: wait.ForHTTP("/").WithStartupTimeout(1 * time.Second),
		},
		Started: true,
	})
	terminateContainerOnEnd(t, ctx, nginxC)

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerCreationWaitsForLogContextTimeout(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("test context timeout").WithStartupTimeout(1 * time.Second),
	}
	_, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerCreationWaitsForLog(t *testing.T) {
	// exposePorts {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
	}
	mysqlC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})
	// }

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, mysqlC)

	// containerHost {
	host, _ := mysqlC.Host(ctx)
	// }
	// mappedPort {
	p, _ := mysqlC.MappedPort(ctx, "3306/tcp")
	port := p.Int()
	// }
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
		"root", "password", host, port, "database")

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		" PRIMARY KEY (`col_1`, `col_2`) \n" +
		")")
	if err != nil {
		t.Errorf("error creating table: %+v\n", err)
	}
}

func Test_BuildContainerFromDockerfile(t *testing.T) {
	t.Log("getting context")
	ctx := context.Background()
	t.Log("got context, creating container request")
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context: "./testresources",
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisC)

	checkSuccessfulRedisImage(ctx, redisC, t)

}

func Test_BuildContainerFromDockerfileWithAuthConfig_ShouldSucceedWithAuthConfigs(t *testing.T) {
	prepareLocalRegistryWithAuth(t)
	defer func() {
		ctx := context.Background()
		testcontainersClient, err := client.NewClientWithOpts(client.WithVersion(daemonMaxVersion))
		if err != nil {
			t.Log("could not create client to cleanup registry: ", err)
		}

		_, err = testcontainersClient.ImageRemove(ctx, "localhost:5000/redis:5.0-alpine", types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Log("could not remove image: ", err)
		}

	}()

	t.Log("getting context")
	ctx := context.Background()
	t.Log("got context, creating container request")
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources",
			Dockerfile: "auth.Dockerfile",
			AuthConfigs: map[string]types.AuthConfig{
				"localhost:5000": {
					Username: "testuser",
					Password: "testpassword",
				},
			},
		},

		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, redisC)

	checkSuccessfulRedisImage(ctx, redisC, t)
}

func Test_BuildContainerFromDockerfileWithAuthConfig_ShouldFailWithoutAuthConfigs(t *testing.T) {
	prepareLocalRegistryWithAuth(t)

	t.Log("getting context")
	ctx := context.Background()
	t.Log("got context, creating container request")
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources",
			Dockerfile: "auth.Dockerfile",
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := prepareRedisImage(ctx, req, t)
	require.Error(t, err)
	terminateContainerOnEnd(t, ctx, redisC)
}

func prepareLocalRegistryWithAuth(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	req := ContainerRequest{
		Image:        "registry:2",
		ExposedPorts: []string{"5000:5000/tcp"},
		Env: map[string]string{
			"REGISTRY_AUTH":                             "htpasswd",
			"REGISTRY_AUTH_HTPASSWD_REALM":              "Registry",
			"REGISTRY_AUTH_HTPASSWD_PATH":               "/auth/htpasswd",
			"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY": "/data",
		},
		Mounts: ContainerMounts{
			ContainerMount{
				Source: GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/testresources/auth", wd),
				},
				Target: "/auth",
			},
			ContainerMount{
				Source: GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/testresources/data", wd),
				},
				Target: "/data",
			},
		},
		WaitingFor: wait.ForExposedPort(),
	}

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	t.Log("creating registry container")

	registryC, err := GenericContainer(ctx, genContainerReq)
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, registryC.Terminate(context.Background()))
	})

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
}

func prepareRedisImage(ctx context.Context, req ContainerRequest, t *testing.T) (Container, error) {
	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	t.Log("creating redis container")

	redisC, err := GenericContainer(ctx, genContainerReq)

	t.Log("created redis container")

	return redisC, err
}

func checkSuccessfulRedisImage(ctx context.Context, redisC Container, t *testing.T) {
	t.Log("created redis container")

	t.Log("getting redis container endpoint")
	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("retrieved redis container endpoint")

	redisClient := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	t.Log("pinging redis")
	pong, err := redisClient.Ping(ctx).Result()
	require.NoError(t, err)

	t.Log("received response from redis")

	if pong != "PONG" {
		t.Fatalf("received unexpected response from redis: %s", pong)
	}
}

func Test_BuildContainerFromDockerfileWithBuildArgs(t *testing.T) {
	t.Log("getting ctx")
	ctx := context.Background()

	ba := "build args value"

	t.Log("got ctx, creating container request")
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources",
			Dockerfile: "args.Dockerfile",
			BuildArgs: map[string]*string{
				"FOO": &ba,
			},
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
	}

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, genContainerReq)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	ep, err := c.Endpoint(ctx, "http")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(ep + "/env")
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, ba, string(body))
}

func Test_BuildContainerFromDockerfileWithBuildLog(t *testing.T) {
	rescueStdout := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	t.Log("getting ctx")
	ctx := context.Background()
	t.Log("got ctx, creating container request")

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:       "./testresources",
			Dockerfile:    "buildlog.Dockerfile",
			PrintBuildLog: true,
		},
	}

	genContainerReq := GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, genContainerReq)

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	temp := strings.Split(string(out), "\n")

	if !regexp.MustCompile(`(?i)^Step\s*1/1\s*:\s*FROM docker.io/alpine$`).MatchString(temp[0]) {
		t.Errorf("Expected stdout firstline to be %s. Got '%s'.", "Step 1/1 : FROM docker.io/alpine", temp[0])
	}
}

func TestContainerCreationWaitsForLogAndPortContextTimeout(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("I love testcontainers-go"),
			wait.ForListeningPort("3306/tcp"),
		),
	}
	_, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	if err == nil {
		t.Fatal("Expected timeout")
	}
}

func TestContainerCreationWaitingForHostPort(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
	}
	nginx, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginx)
}

func TestContainerCreationWaitingForHostPortWithoutBashThrowsAnError(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
	}
	nginx, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginx)
}

func TestContainerCreationWaitsForLogAndPort(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("port: 3306  MySQL Community Server - GPL"),
			wait.ForListeningPort("3306/tcp"),
		),
	}

	mysqlC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, mysqlC)

	// buildingAddresses {
	host, _ := mysqlC.Host(ctx)
	p, _ := mysqlC.MappedPort(ctx, "3306/tcp")
	port := p.Int()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
		"root", "password", host, port, "database")

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	// }

	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
}

func TestCMD(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in a command to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := ContainerRequest{
		Image: "docker.io/alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("command override!"),
		),
		Cmd: []string{"echo", "command override!"},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)
}

func TestEntrypoint(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in an entrypoint to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := ContainerRequest{
		Image: "docker.io/alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("entrypoint override!"),
		),
		Entrypoint: []string{"echo", "entrypoint override!"},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)
}

func TestReadTCPropsFile(t *testing.T) {
	t.Run("HOME is not set", func(t *testing.T) {
		t.Setenv("HOME", "")

		config := configureTC()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME is not set - TESTCONTAINERS_ env is set", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := configureTC()

		expected := TestContainersConfig{}
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME does not contain TC props file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		config := configureTC()

		assert.Empty(t, config, "TC props file should not exist")
	})

	t.Run("HOME does not contain TC props file - TESTCONTAINERS_ env is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")

		config := configureTC()
		expected := TestContainersConfig{}
		expected.RyukPrivileged = true

		assert.Equal(t, expected, config)
	})

	t.Run("HOME contains TC properties file", func(t *testing.T) {
		tests := []struct {
			content  string
			env      map[string]string
			expected TestContainersConfig
		}{
			{
				"docker.host = tcp://127.0.0.1:33293",
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"docker.host = tcp://127.0.0.1:33293",
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	`,
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:4711",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	docker.host = tcp://127.0.0.1:1234
	docker.tls.verify = 1
	`,
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 1,
					CertPath:  "",
				},
			},
			{
				"",
				map[string]string{},
				TestContainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`foo = bar
	docker.host = tcp://127.0.0.1:1234
			`,
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				"docker.host=tcp://127.0.0.1:33293",
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:33293",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`#docker.host=tcp://127.0.0.1:33293`,
				map[string]string{},
				TestContainersConfig{
					Host:      "",
					TLSVerify: 0,
					CertPath:  "",
				},
			},
			{
				`#docker.host = tcp://127.0.0.1:33293
	docker.host = tcp://127.0.0.1:4711
	docker.host = tcp://127.0.0.1:1234
	docker.cert.path=/tmp/certs`,
				map[string]string{},
				TestContainersConfig{
					Host:      "tcp://127.0.0.1:1234",
					TLSVerify: 0,
					CertPath:  "/tmp/certs",
				},
			},
			{
				`ryuk.container.privileged=true`,
				map[string]string{},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				``,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				`ryuk.container.privileged=true`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
			},
			{
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "false",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
			},
			{
				`ryuk.container.privileged=false
				docker.tls.verify = ERROR`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "true",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: true,
				},
			},
			{
				`ryuk.container.privileged=false`,
				map[string]string{
					"TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED": "foo",
				},
				TestContainersConfig{
					Host:           "",
					TLSVerify:      0,
					CertPath:       "",
					RyukPrivileged: false,
				},
			},
		}
		for i, tt := range tests {
			t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
				tmpDir := t.TempDir()
				t.Setenv("HOME", tmpDir)
				for k, v := range tt.env {
					t.Setenv(k, v)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, ".testcontainers.properties"), []byte(tt.content), 0o600); err != nil {
					t.Errorf("Failed to create the file: %v", err)
					return
				}

				config := configureTC()

				assert.Equal(t, tt.expected, config, "Configuration doesn't not match")

			})
		}
	})
}

func ExampleDockerProvider_CreateContainer() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
}

func ExampleContainer_Host() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	ip, _ := nginxC.Host(ctx)
	println(ip)
}

func ExampleContainer_Start() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	_ = nginxC.Start(ctx)
}

func ExampleContainer_Stop() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	timeout := 10 * time.Second
	_ = nginxC.Stop(ctx, &timeout)
}

func ExampleContainer_MappedPort() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	ip, _ := nginxC.Host(ctx)
	port, _ := nginxC.MappedPort(ctx, "80")
	_, _ = http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
}

func TestContainerCreationWithBindAndVolume(t *testing.T) {
	absPath, err := filepath.Abs("./testresources/hello.sh")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()
	// Create a Docker client.
	dockerCli, _, _, err := NewDockerClient()
	if err != nil {
		t.Fatal(err)
	}

	// Create the volume.
	vol, err := dockerCli.VolumeCreate(ctx, volume.CreateOptions{
		Driver: "local",
	})
	if err != nil {
		t.Fatal(err)
	}
	volumeName := vol.Name
	t.Cleanup(func() {
		ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cnl()
		err := dockerCli.VolumeRemove(ctx, volumeName, true)
		if err != nil {
			t.Fatal(err)
		}
	})
	// Create the container that writes into the mounted volume.
	bashC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:      "docker.io/bash",
			Mounts:     Mounts(BindMount(absPath, "/hello.sh"), VolumeMount(volumeName, "/data")),
			Cmd:        []string{"bash", "/hello.sh"},
			WaitingFor: wait.ForLog("done"),
		},
		Started: true,
	})

	require.NoError(t, err)
	require.NoError(t, bashC.Terminate(ctx))
}

func TestContainerWithTmpFs(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: "docker.io/busybox",
		Cmd:   []string{"sleep", "10"},
		Tmpfs: map[string]string{"/testtmpfs": "rw"},
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	path := "/testtmpfs/test.file"

	c, _, err := container.Exec(ctx, []string{"ls", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 1 {
		t.Fatalf("File %s should not have existed, expected return code 1, got %v", path, c)
	}

	c, _, err = container.Exec(ctx, []string{"touch", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should have been created successfully, expected return code 0, got %v", path, c)
	}

	c, _, err = container.Exec(ctx, []string{"ls", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", path, c)
	}
}

func TestContainerNonExistentImage(t *testing.T) {
	t.Run("if the image not found don't propagate the error", func(t *testing.T) {
		_, err := GenericContainer(context.Background(), GenericContainerRequest{
			ContainerRequest: ContainerRequest{
				Image:      "postgres:nonexistent-version",
				SkipReaper: true,
			},
			Started: true,
		})

		var nf errdefs.ErrNotFound
		if !errors.As(err, &nf) {
			t.Fatalf("the error should have bee an errdefs.ErrNotFound: %v", err)
		}
	})

	t.Run("the context cancellation is propagated to container creation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := GenericContainer(ctx, GenericContainerRequest{
			ProviderType: providerType,
			ContainerRequest: ContainerRequest{
				Image:      "docker.io/postgres:12",
				WaitingFor: wait.ForLog("log"),
				SkipReaper: true,
			},
			Started: true,
		})
		if !errors.Is(err, ctx.Err()) {
			t.Fatalf("err should be a ctx cancelled error %v", err)
		}
	})
}

func TestContainerCustomPlatformImage(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Incompatible Docker API version for Podman")
	}
	t.Run("error with a non-existent platform", func(t *testing.T) {
		t.Parallel()
		nonExistentPlatform := "windows/arm12"
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		c, err := GenericContainer(ctx, GenericContainerRequest{
			ProviderType: providerType,
			ContainerRequest: ContainerRequest{
				Image:         "docker.io/redis:latest",
				SkipReaper:    true,
				ImagePlatform: nonExistentPlatform,
			},
			Started: false,
		})

		terminateContainerOnEnd(t, ctx, c)

		assert.Error(t, err)
	})

	t.Run("specific platform should be propagated", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, err := GenericContainer(ctx, GenericContainerRequest{
			ProviderType: providerType,
			ContainerRequest: ContainerRequest{
				Image:         "docker.io/mysql:5.7",
				SkipReaper:    true,
				ImagePlatform: "linux/amd64",
			},
			Started: false,
		})

		require.NoError(t, err)
		terminateContainerOnEnd(t, ctx, c)

		dockerCli, _, _, err := NewDockerClient()
		require.NoError(t, err)

		dockerCli.NegotiateAPIVersion(ctx)
		ctr, err := dockerCli.ContainerInspect(ctx, c.GetContainerID())
		assert.NoError(t, err)

		img, _, err := dockerCli.ImageInspectWithRaw(ctx, ctr.Image)
		assert.NoError(t, err)
		assert.Equal(t, "linux", img.Os)
		assert.Equal(t, "amd64", img.Architecture)
	})
}

func TestContainerWithCustomHostname(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("some-nginx-%s-%d", t.Name(), rand.Int())
	hostname := fmt.Sprintf("my-nginx-%s-%d", t.Name(), rand.Int())
	req := ContainerRequest{
		Name:     name,
		Image:    nginxImage,
		Hostname: hostname,
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	if actualHostname := readHostname(t, container.GetContainerID()); actualHostname != hostname {
		t.Fatalf("expected hostname %s, got %s", hostname, actualHostname)
	}
}

func readHostname(tb testing.TB, containerId string) string {
	containerClient, _, _, err := NewDockerClient()
	if err != nil {
		tb.Fatalf("Failed to create Docker client: %v", err)
	}

	containerDetails, err := containerClient.ContainerInspect(context.Background(), containerId)
	if err != nil {
		tb.Fatalf("Failed to inspect container: %v", err)
	}

	return containerDetails.Config.Hostname
}

func TestDockerContainerCopyFileToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	copiedFileName := "hello_copy.sh"
	_ = nginxC.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}
}

func TestDockerContainerCopyDirToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	err = nginxC.CopyDirToContainer(ctx, "./testresources/Dockerfile", "/tmp/testresources/Dockerfile", 700)
	require.Error(t, err) // copying a file using the directory method will raise an error

	err = nginxC.CopyDirToContainer(ctx, "./testresources", "/tmp/testresources", 700)
	if err != nil {
		t.Fatal(err)
	}

	assertExtractedFiles(t, ctx, nginxC, "./testresources", "/tmp/testresources/")
}

func TestDockerCreateContainerWithFiles(t *testing.T) {
	ctx := context.Background()
	hostFileName := "./testresources/hello.sh"
	copiedFileName := "/hello_copy.sh"
	tests := []struct {
		name   string
		files  []ContainerFile
		errMsg string
	}{
		{
			name: "success copy",
			files: []ContainerFile{
				{
					HostFilePath:      hostFileName,
					ContainerFilePath: copiedFileName,
					FileMode:          700,
				},
			},
		},
		{
			name: "host file not found",
			files: []ContainerFile{
				{
					HostFilePath:      hostFileName + "123",
					ContainerFilePath: copiedFileName,
					FileMode:          700,
				},
			},
			errMsg: "can't copy " +
				"./testresources/hello.sh123 to container: open " +
				"./testresources/hello.sh123: no such file or directory: " +
				"failed to create container",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: ContainerRequest{
					Image:        "nginx:1.17.6",
					ExposedPorts: []string{"80/tcp"},
					WaitingFor:   wait.ForListeningPort("80/tcp"),
					Files:        tc.files,
				},
				Started: false,
			})

			if err != nil {
				require.Contains(t, err.Error(), tc.errMsg)
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
	hostDirName := "testresources"

	tests := []struct {
		name     string
		dir      ContainerFile
		hasError bool
	}{
		{
			name: "success copy directory",
			dir: ContainerFile{
				HostFilePath:      "./" + hostDirName,
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          700,
			},
			hasError: false,
		},
		{
			name: "host dir not found",
			dir: ContainerFile{
				HostFilePath:      "./testresources123",  // does not exist
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          700,
			},
			hasError: true,
		},
		{
			name: "container dir not found",
			dir: ContainerFile{
				HostFilePath:      "./" + hostDirName,
				ContainerFilePath: "/parent-does-not-exist/testresources123", // does not exist
				FileMode:          700,
			},
			hasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: ContainerRequest{
					Image:        "nginx:1.17.6",
					ExposedPorts: []string{"80/tcp"},
					WaitingFor:   wait.ForListeningPort("80/tcp"),
					Files:        []ContainerFile{tc.dir},
				},
				Started: false,
			})

			require.True(t, (err != nil) == tc.hasError)
			if err == nil {
				dir := tc.dir

				assertExtractedFiles(t, ctx, nginxC, dir.HostFilePath, dir.ContainerFilePath)
			}
		})
	}
}

func TestDockerContainerCopyToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	copiedFileName := "hello_copy.sh"

	fileContent, err := os.ReadFile("./testresources/hello.sh")
	if err != nil {
		t.Fatal(err)
	}
	_ = nginxC.CopyToContainer(ctx, fileContent, "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}
}

func TestDockerContainerCopyFileFromContainer(t *testing.T) {
	fileContent, err := os.ReadFile("./testresources/hello.sh")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	copiedFileName := "hello_copy.sh"
	_ = nginxC.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	if err != nil {
		t.Fatal(err)
	}

	fileContentFromContainer, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fileContent, fileContentFromContainer)
}

func TestDockerContainerCopyEmptyFileFromContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	copiedFileName := "hello_copy.sh"
	_ = nginxC.CopyFileToContainer(ctx, "./testresources/empty.sh", "/"+copiedFileName, 700)
	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	if err != nil {
		t.Fatal(err)
	}

	fileContentFromContainer, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, fileContentFromContainer)
}

func TestDockerContainerResources(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Rootless Podman does not support setting rlimit")
	}

	ctx := context.Background()

	expected := []*units.Ulimit{
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

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
			Resources: container.Resources{
				Ulimits: expected,
			},
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	c, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	c.NegotiateAPIVersion(ctx)
	containerID := nginxC.GetContainerID()

	resp, err := c.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	assert.Equal(t, expected, resp.HostConfig.Ulimits)
}

func TestContainerWithReaperNetwork(t *testing.T) {
	ctx := context.Background()
	networks := []string{
		"test_network_" + randomString(),
		"test_network_" + randomString(),
	}

	for _, nw := range networks {
		nr := NetworkRequest{
			Name:       nw,
			Attachable: true,
		}
		_, err := GenericNetwork(ctx, GenericNetworkRequest{
			ProviderType:   providerType,
			NetworkRequest: nr,
		})
		assert.Nil(t, err)
	}

	req := ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nginxDefaultPort),
			wait.ForLog("Configuration complete; ready for start up"),
		),
		Networks: networks,
	}

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxC)

	containerId := nginxC.GetContainerID()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err)
	cnt, err := cli.ContainerInspect(ctx, containerId)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cnt.NetworkSettings.Networks))
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[0]])
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[1]])
}

func TestContainerCapAdd(t *testing.T) {
	if providerType == ProviderPodman {
		t.Skip("Rootless Podman does not support setting cap-add/cap-drop")
	}

	ctx := context.Background()

	expected := "IPC_LOCK"

	nginx, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
			CapAdd:       []string{expected},
		},
		Started: true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginx)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)
	defer dockerClient.Close()

	containerID := nginx.GetContainerID()
	resp, err := dockerClient.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	assert.Equal(t, strslice.StrSlice{expected}, resp.HostConfig.CapAdd)
}

func TestContainerRunningCheckingStatusCode(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "influxdb:1.8.10-alpine",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/ping").WithPort("8086/tcp").WithStatusCodeMatcher(
				func(status int) bool {
					return status == http.StatusNoContent
				},
			),
		),
	}

	influx, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	terminateContainerOnEnd(t, ctx, influx)
}

func TestContainerWithUserID(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:      "docker.io/alpine:latest",
		User:       "60125",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	r, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	actual := regexp.MustCompile(`\D+`).ReplaceAllString(string(b), "")
	assert.Equal(t, req.User, actual)
}

func TestContainerWithNoUserID(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:      "docker.io/alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	r, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	actual := regexp.MustCompile(`\D+`).ReplaceAllString(string(b), "")
	assert.Equal(t, "0", actual)
}

func TestGetGatewayIP(t *testing.T) {
	// When using docker-compose with DinD mode, and using host port or http wait strategy
	// It's need to invoke GetGatewayIP for get the host
	provider, err := providerType.GetProvider(WithLogger(TestLogger(t)))
	if err != nil {
		t.Fatal(err)
	}
	ip, err := provider.(*DockerProvider).GetGatewayIP(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ip == "" {
		t.Fatal("could not get gateway ip")
	}
}

func TestProviderHasConfig(t *testing.T) {
	provider, err := NewDockerProvider(WithLogger(TestLogger(t)))
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, provider.Config(), "expecting DockerProvider to provide the configuration")
}

func TestNetworkModeWithContainerReference(t *testing.T) {
	ctx := context.Background()
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxA)

	networkMode := fmt.Sprintf("container:%v", nginxA.GetContainerID())
	nginxB, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:       nginxAlpineImage,
			NetworkMode: container.NetworkMode(networkMode),
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxB)
}

// creates a temporary dir in which the files will be extracted. Then it will compare the bytes of each file in the source with the bytes from the copied-from-container file
func assertExtractedFiles(t *testing.T, ctx context.Context, container Container, hostFilePath string, containerFilePath string) {
	// create all copied files into a temporary dir
	tmpDir := filepath.Join(t.TempDir())

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

		// copy file by file, as there is a limitation in the Docker client to copy an entiry directory from the container
		// paths for the container files are using Linux path separators
		fd, err := container.CopyFileFromContainer(ctx, containerFilePath+"/"+srcFile.Name())
		require.NoError(t, err, "Path not found in container: %s", containerFilePath+"/"+srcFile.Name())
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
		assert.Equal(t, srcBytes, untarBytes)
	}
}

func terminateContainerOnEnd(tb testing.TB, ctx context.Context, ctr Container) {
	tb.Helper()
	if ctr == nil {
		return
	}
	tb.Cleanup(func() {
		tb.Log("terminating container")
		require.NoError(tb, ctr.Terminate(ctx))
	})
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func TestDockerProviderFindContainerByName(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDockerProvider(WithLogger(TestLogger(t)))
	require.NoError(t, err)

	c1, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Name:       "test",
			Image:      "nginx:1.17.6",
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	})
	require.NoError(t, err)
	c1Name, err := c1.Name(ctx)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c1)

	c2, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Name:       "test2",
			Image:      "nginx:1.17.6",
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c2)

	c, err := provider.findContainerByName(ctx, "test")
	assert.NoError(t, err)
	require.NotNil(t, c)
	assert.Contains(t, c.Names, c1Name)
}
