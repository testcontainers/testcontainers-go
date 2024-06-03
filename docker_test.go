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
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-units"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tcimage "github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/internal/core"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestContainerWithHostNetworkOptions(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)

	absPath, err := filepath.Abs(filepath.Join("testdata", "nginx-highport.conf"))
	if err != nil {
		t.Fatal(err)
	}

	req := Request{
		Image: nginxAlpineImage,
		Files: []ContainerFile{
			{
				HostFilePath:      absPath,
				ContainerFilePath: "/etc/nginx/conf.d/default.conf",
			},
		},
		ExposedPorts: []string{
			nginxHighPort,
		},
		Privileged: true,
		WaitingFor: wait.ForListeningPort(nginxHighPort),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.NetworkMode = "host"
		},
		Started: true,
	}

	nginxC, err := New(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

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

func TestContainer_UseExposePortsFromImageConfigs(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:      "nginx",
		Privileged: true,
		WaitingFor: wait.ForExposedPort(),
		Started:    true,
	}

	nginxC, err := New(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	TerminateContainerOnEnd(t, ctx, nginxC)

	endpoint, err := nginxC.Endpoint(ctx, "http")
	if err != nil {
		t.Errorf("Expected server endpoint. Got '%v'.", err)
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerWithNetworkModeAndNetworkTogether(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	// skipIfDockerDesktop {
	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)
	// }

	req := Request{
		Image:    nginxImage,
		Networks: []string{"new-network"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.NetworkMode = "host"
		},
		Started: true,
	}

	nginx, err := New(ctx, req)
	if err != nil {
		// Error when NetworkMode = host and Network = []string{"bridge"}
		t.Logf("Can't use Network and NetworkMode together, %s\n", err)
	}
	TerminateContainerOnEnd(t, ctx, nginx)
}

func TestContainerWithHostNetwork(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping test that requires host network access when running in a container")
	}

	ctx := context.Background()
	SkipIfDockerDesktop(t, ctx)

	absPath, err := filepath.Abs(filepath.Join("testdata", "nginx-highport.conf"))
	if err != nil {
		t.Fatal(err)
	}

	req := Request{
		Image:      nginxAlpineImage,
		WaitingFor: wait.ForListeningPort(nginxHighPort),
		Files: []ContainerFile{
			{
				HostFilePath:      absPath,
				ContainerFilePath: "/etc/nginx/conf.d/default.conf",
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.NetworkMode = "host"
		},
		Started: true,
	}

	nginxC, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	portEndpoint, err := nginxC.PortEndpoint(ctx, nginxHighPort, "http")
	if err != nil {
		t.Errorf("Expected port endpoint %s. Got '%d'.", portEndpoint, err)
	}
	t.Log(portEndpoint)

	_, err = http.Get(portEndpoint)
	if err != nil {
		t.Errorf("Expected OK response. Got '%v'.", err)
	}

	host, err := nginxC.Host(ctx)
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", host, err)
	}

	_, err = http.Get("http://" + host + ":8080")
	if err != nil {
		t.Errorf("Expected OK response. Got '%v'.", err)
	}
}

func TestContainerReturnItsContainerID(t *testing.T) {
	ctx := context.Background()
	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxA)

	if nginxA.GetContainerID() == "" {
		t.Errorf("expected a containerID but we got an empty string.")
	}
}

func TestContainerTerminationResetsState(t *testing.T) {
	ctx := context.Background()

	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
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
	if nginxA.SessionID() != "" {
		t.Fatal("Internal state must be reset.")
	}
	inspect, err := nginxA.Inspect(ctx)
	if err == nil || inspect != nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerStateAfterTermination(t *testing.T) {
	createContainerFn := func(ctx context.Context) (StartedContainer, error) {
		return New(ctx, Request{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
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
		require.Error(t, err, "expected error from container inspect.")

		assert.Nil(t, state, "expected nil container inspect.")
	})

	t.Run("Nil State after termination if raw as already set", func(t *testing.T) {
		ctx := context.Background()
		nginx, err := createContainerFn(ctx)
		if err != nil {
			t.Fatal(err)
		}

		state, err := nginx.State(ctx)
		require.NoError(t, err, "unexpected error from container inspect before container termination.")

		assert.NotNil(t, state, "unexpected nil container inspect before container termination.")

		// terminate the container before the raw state is set
		err = nginx.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}

		state, err = nginx.State(ctx)
		require.Error(t, err, "expected error from container inspect after container termination.")

		assert.Nil(t, state, "unexpected nil container inspect after container termination.")
	})
}

func TestContainerTerminationRemovesDockerImage(t *testing.T) {
	t.Run("if not built from Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		dockerClient, err := core.NewClient(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer dockerClient.Close()

		ctr, err := New(ctx, Request{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			Started: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = ctr.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = dockerClient.ImageInspectWithRaw(ctx, nginxAlpineImage)
		if err != nil {
			t.Fatal("nginx image should not have been removed")
		}
	})

	t.Run("if built from Dockerfile", func(t *testing.T) {
		ctx := context.Background()
		dockerClient, err := core.NewClient(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer dockerClient.Close()

		ctr, err := New(ctx, Request{
			FromDockerfile: FromDockerfile{
				Context: filepath.Join(".", "testdata"),
			},
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
			Started:      true,
		})
		if err != nil {
			t.Fatal(err)
		}
		containerID := ctr.GetContainerID()
		resp, err := dockerClient.ContainerInspect(ctx, containerID)
		if err != nil {
			t.Fatal(err)
		}
		imageID := resp.Config.Image

		err = ctr.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = dockerClient.ImageInspectWithRaw(ctx, imageID)
		if err == nil {
			t.Fatal("custom built image should have been removed", err)
		}
	})
}

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxA)

	nginxB, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxB)

	endpointA, err := nginxA.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpointA)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()

	nginxC, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

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

func TestContainerCreationWithName(t *testing.T) {
	ctx := context.Background()

	creationName := fmt.Sprintf("%s_%d", "test_container", time.Now().Unix())
	expectedName := "/" + creationName // inspect adds '/' in the beginning

	nginxC, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		Name:       creationName,
		Networks:   []string{"bridge"},
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	inspect, err := nginxC.Inspect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	name := inspect.Name
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
	if network != corenetwork.Bridge {
		t.Errorf("Expected network name '%s'. Got '%s'.", corenetwork.Bridge, network)
	}

	endpoint, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	require.NoError(t, err)

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationAndWaitForListeningPortLongEnough(t *testing.T) {
	ctx := context.Background()

	// delayed-nginx will wait 2s before opening port
	nginxC, err := New(ctx, Request{
		Image: nginxDelayedImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort), // default startupTimeout is 60s
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(origin)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationTimesOut(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := New(ctx, Request{
		Image: nginxDelayedImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort).WithStartupTimeout(1 * time.Second),
		Started:    true,
	})

	TerminateContainerOnEnd(t, ctx, nginxC)

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerRespondsWithHttp200ForIndex(t *testing.T) {
	ctx := context.Background()

	nginxC, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	origin, err := nginxC.PortEndpoint(ctx, nginxDefaultPort, "http")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(origin)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationTimesOutWithHttp(t *testing.T) {
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := New(ctx, Request{
		Image: nginxDelayedImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		WaitingFor: wait.ForHTTP("/").WithStartupTimeout(1 * time.Second),
		Started:    true,
	})
	TerminateContainerOnEnd(t, ctx, nginxC)

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerCreationWaitsForLogContextTimeout(t *testing.T) {
	ctx := context.Background()

	c, err := New(ctx, Request{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("test context timeout").WithStartupTimeout(1 * time.Second),
		Started:    true,
	})
	if err == nil {
		t.Error("Expected timeout")
	}

	TerminateContainerOnEnd(t, ctx, c)
}

func TestContainerCreationWaitsForLog(t *testing.T) {
	ctx := context.Background()

	mysqlC, err := New(ctx, Request{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
		Started:    true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, mysqlC)
}

func TestBuildContainerFromDockerfileWithBuildArgs(t *testing.T) {
	t.Log("getting ctx")
	ctx := context.Background()

	t.Log("got ctx, creating container request")

	// fromDockerfileWithBuildArgs {
	ba := "build args value"

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    filepath.Join(".", "testdata"),
			Dockerfile: "args.Dockerfile",
			BuildArgs: map[string]*string{
				"FOO": &ba,
			},
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		Started:      true,
	}
	// }

	c, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)

	ep, err := c.Endpoint(ctx, "http")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(ep + "/env")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, ba, string(body))
}

func TestBuildContainerFromDockerfileWithBuildLog(t *testing.T) {
	rescueStdout := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	t.Log("getting ctx")
	ctx := context.Background()
	t.Log("got ctx, creating container request")

	// fromDockerfile {
	req := Request{
		FromDockerfile: FromDockerfile{
			Context:       filepath.Join(".", "testdata"),
			Dockerfile:    "buildlog.Dockerfile",
			PrintBuildLog: true,
		},
		Started: true,
	}
	// }

	c, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout
	temp := strings.Split(string(out), "\n")

	if !regexp.MustCompile(`^Step\s*1/\d+\s*:\s*FROM docker.io/alpine$`).MatchString(temp[0]) {
		t.Errorf("Expected stdout firstline to be %s. Got '%s'.", "Step 1/* : FROM docker.io/alpine", temp[0])
	}
}

func TestContainerCreationWaitsForLogAndPortContextTimeout(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:        mysqlImage,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("I love testcontainers-go"),
			wait.ForListeningPort("3306/tcp"),
		).WithDeadline(5 * time.Second),
		Started: true,
	}
	c, err := New(ctx, req)
	if err == nil {
		t.Fatal("Expected timeout")
	}

	TerminateContainerOnEnd(t, ctx, c)
}

func TestContainerCreationWaitingForHostPort(t *testing.T) {
	ctx := context.Background()
	// exposePorts {
	req := Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		Started:      true,
	}
	// }
	nginx, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginx)
}

func TestContainerCreationWaitingForHostPortWithoutBashThrowsAnError(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		Started:      true,
	}
	nginx, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginx)
}

func TestCMD(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in a command to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := Request{
		Image: "docker.io/alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("command override!"),
		),
		Cmd:     []string{"echo", "command override!"},
		Started: true,
	}

	c, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)
}

func TestEntrypoint(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in an entrypoint to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := Request{
		Image: "docker.io/alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("entrypoint override!"),
		),
		Entrypoint: []string{"echo", "entrypoint override!"},
		Started:    true,
	}

	c, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)
}

func TestWorkingDir(t *testing.T) {
	/*
		print the current working directory to ensure that
		we can specify working directory in the
		ContainerRequest and it will be used for the container
	*/

	ctx := context.Background()

	req := Request{
		Image: "docker.io/alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("/var/tmp/test"),
		),
		Entrypoint: []string{"pwd"},
		WorkingDir: "/var/tmp/test",
		Started:    true,
	}

	c, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)
}

func TestContainerCreationWithVolumeAndFileWritingToIt(t *testing.T) {
	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Create the volume.
	volumeName := "volumeName"

	// Create the container that writes into the mounted volume.
	bashC, err := New(ctx, Request{
		Image: "docker.io/bash",
		Files: []ContainerFile{
			{
				HostFilePath:      absPath,
				ContainerFilePath: "/hello.sh",
			},
		},
		Mounts:     tcmount.Mounts(tcmount.VolumeMount(volumeName, "/data")),
		Cmd:        []string{"bash", "/hello.sh"},
		WaitingFor: wait.ForLog("done"),
		Started:    true,
	})

	require.NoError(t, err)
	require.NoError(t, bashC.Terminate(ctx))
}

func TestContainerWithTmpFs(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:   "docker.io/busybox",
		Cmd:     []string{"sleep", "10"},
		Tmpfs:   map[string]string{"/testtmpfs": "rw"},
		Started: true,
	}

	ctr, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, ctr)

	path := "/testtmpfs/test.file"

	// exec_reader_example {
	c, reader, err := ctr.Exec(ctx, []string{"ls", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 1 {
		t.Fatalf("File %s should not have existed, expected return code 1, got %v", path, c)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		t.Fatal(err)
	}

	// See the logs from the command execution.
	t.Log(buf.String())
	// }

	// exec_example {
	c, _, err = ctr.Exec(ctx, []string{"touch", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should have been created successfully, expected return code 0, got %v", path, c)
	}
	// }

	c, _, err = ctr.Exec(ctx, []string{"ls", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", path, c)
	}
}

func TestContainerNonExistentImage(t *testing.T) {
	t.Run("if the image not found don't propagate the error", func(t *testing.T) {
		_, err := New(context.Background(), Request{
			Image:   "postgres:nonexistent-version",
			Started: true,
		})

		var nf errdefs.ErrNotFound
		if !errors.As(err, &nf) {
			t.Fatalf("the error should have been an errdefs.ErrNotFound: %v", err)
		}
	})

	t.Run("the context cancellation is propagated to container creation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		c, err := New(ctx, Request{
			Image:      "docker.io/postgres:12",
			WaitingFor: wait.ForLog("log"),
			Started:    true,
		})
		if !errors.Is(err, ctx.Err()) {
			t.Fatalf("err should be a ctx cancelled error %v", err)
		}

		TerminateContainerOnEnd(t, context.Background(), c) // use non-cancelled context
	})
}

func TestContainerCustomPlatformImage(t *testing.T) {
	t.Run("error with a non-existent platform", func(t *testing.T) {
		t.Parallel()
		nonExistentPlatform := "windows/arm12"
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		c, err := New(ctx, Request{
			Image:         "docker.io/redis:latest",
			ImagePlatform: nonExistentPlatform,
			Started:       false,
		})

		TerminateContainerOnEnd(t, ctx, c)

		require.Error(t, err)
	})

	t.Run("specific platform should be propagated", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, err := New(ctx, Request{
			Image:         "docker.io/mysql:8.0.36",
			ImagePlatform: "linux/amd64",
			Started:       false,
		})

		require.NoError(t, err)
		TerminateContainerOnEnd(t, ctx, c)

		dockerCli, err := core.NewClient(ctx)
		require.NoError(t, err)
		defer dockerCli.Close()

		ctr, err := dockerCli.ContainerInspect(ctx, c.GetContainerID())
		require.NoError(t, err)

		img, _, err := dockerCli.ImageInspectWithRaw(ctx, ctr.Image)
		require.NoError(t, err)
		assert.Equal(t, "linux", img.Os)
		assert.Equal(t, "amd64", img.Architecture)
	})
}

func TestContainerWithCustomHostname(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("some-nginx-%s-%d", t.Name(), rand.Int())
	hostname := fmt.Sprintf("my-nginx-%s-%d", t.Name(), rand.Int())
	req := Request{
		Name:     name,
		Image:    nginxImage,
		Hostname: hostname,
		Started:  true,
	}
	ctr, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, ctr)

	if actualHostname := readHostname(t, ctr.GetContainerID()); actualHostname != hostname {
		t.Fatalf("expected hostname %s, got %s", hostname, actualHostname)
	}
}

func TestContainerInspect_RawInspectIsCleanedOnStop(t *testing.T) {
	ctr, err := New(context.Background(), Request{
		Image:   nginxImage,
		Started: true,
	})
	require.NoError(t, err)
	TerminateContainerOnEnd(t, context.Background(), ctr)

	inspect, err := ctr.Inspect(context.Background())
	require.NoError(t, err)

	assert.NotEmpty(t, inspect.ID)

	require.NoError(t, ctr.Stop(context.Background(), nil))

	assert.Nil(t, ctr.raw)
}

func readHostname(tb testing.TB, containerId string) string {
	containerClient, err := core.NewClient(context.Background())
	if err != nil {
		tb.Fatalf("Failed to create Docker client: %v", err)
	}
	defer containerClient.Close()

	containerDetails, err := containerClient.ContainerInspect(context.Background(), containerId)
	if err != nil {
		tb.Fatalf("Failed to inspect container: %v", err)
	}

	return containerDetails.Config.Hostname
}

func TestDockerContainerResources(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Rootless Docker does not support setting rlimit")
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

	nginxC, err := New(ctx, Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Resources = container.Resources{
				Ulimits: expected,
			}
		},
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxC)

	c, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer c.Close()

	containerID := nginxC.GetContainerID()

	resp, err := c.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	assert.Equal(t, expected, resp.HostConfig.Ulimits)
}

func TestContainerCapAdd(t *testing.T) {
	ctx := context.Background()

	expected := "IPC_LOCK"

	nginx, err := New(ctx, Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CapAdd = []string{expected}
		},
		Started: true,
	})
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginx)

	dockerClient, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer dockerClient.Close()

	containerID := nginx.GetContainerID()
	resp, err := dockerClient.ContainerInspect(ctx, containerID)
	require.NoError(t, err)

	assert.Equal(t, strslice.StrSlice{expected}, resp.HostConfig.CapAdd)
}

func TestContainerRunningCheckingStatusCode(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:         "influxdb:1.8.10-alpine",
		ExposedPorts:  []string{"8086/tcp"},
		ImagePlatform: "linux/amd64", // influxdb doesn't provide an alpine+arm build (https://github.com/influxdata/influxdata-docker/issues/335)
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/ping").WithPort("8086/tcp").WithStatusCodeMatcher(
				func(status int) bool {
					return status == http.StatusNoContent
				},
			),
		),
		Started: true,
	}

	influx, err := New(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	TerminateContainerOnEnd(t, ctx, influx)
}

func TestContainerWithUserID(t *testing.T) {
	ctx := context.Background()

	req := Request{
		Image:      "docker.io/alpine:latest",
		User:       "60125",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
		Started:    true,
	}

	ctr, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, ctr)

	r, err := ctr.Logs(ctx)
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
	req := Request{
		Image:      "docker.io/alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
		Started:    true,
	}
	ctr, err := New(ctx, req)

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, ctr)

	r, err := ctr.Logs(ctx)
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

func TestNetworkModeWithContainerReference(t *testing.T) {
	ctx := context.Background()
	nginxA, err := New(ctx, Request{
		Image:   nginxAlpineImage,
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxA)

	networkMode := fmt.Sprintf("container:%v", nginxA.GetContainerID())
	nginxB, err := New(ctx, Request{
		Image: nginxAlpineImage,
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.NetworkMode = container.NetworkMode(networkMode)
		},
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxB)
}

func TestFindContainerByName(t *testing.T) {
	ctx := context.Background()

	logger := tclog.NewTestLogger(t)

	c1, err := New(ctx, Request{
		Name:       "test",
		Image:      "nginx:1.17.6",
		Logger:     logger,
		WaitingFor: wait.ForExposedPort(),
		Started:    true,
	})
	require.NoError(t, err)

	c1Inspect, err := c1.Inspect(ctx)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c1)

	c1Name := c1Inspect.Name

	c2, err := New(ctx, Request{
		Name:       "test2",
		Image:      "nginx:1.17.6",
		Logger:     logger,
		WaitingFor: wait.ForExposedPort(),
		Started:    true,
	})
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c2)

	c, err := findContainerByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Contains(t, c.Names, c1Name)
}

func TestImageBuiltFromDockerfile_KeepBuiltImage(t *testing.T) {
	tests := []struct {
		keepBuiltImage bool
	}{
		{keepBuiltImage: true},
		{keepBuiltImage: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Keep built image: %t", tt.keepBuiltImage), func(t *testing.T) {
			ctx := context.Background()

			// Set up CLI.
			cli, err := core.NewClient(ctx)
			require.NoError(t, err, "get docker client should not fail")
			defer func() { _ = cli.Close() }()

			// Create container.
			c, err := New(ctx, Request{
				FromDockerfile: FromDockerfile{
					Context:    "testdata",
					Dockerfile: "echo.Dockerfile",
					KeepImage:  tt.keepBuiltImage,
				},
			})
			require.NoError(t, err, "create container should not fail")
			defer func() { _ = c.Terminate(context.Background()) }()
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
			_, _, err = cli.ImageInspectWithRaw(ctx, containerImage)
			if tt.keepBuiltImage {
				require.NoError(t, err, "image should still exist")
			} else {
				require.Error(t, err, "image should not exist anymore")
			}
		})
	}
}

// errMockCli is a mock implementation of client.APIClient and the client.SystemAPIClient,
// which is handy for simulating error returns in retry scenarios.
type errMockCli struct {
	client.APIClient

	logger             tclog.Logging
	err                error
	imageBuildCount    int
	containerListCount int
	imagePullCount     int
}

func (m *errMockCli) ImageBuild(_ context.Context, _ io.Reader, _ types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.imageBuildCount++
	return types.ImageBuildResponse{Body: io.NopCloser(&bytes.Buffer{})}, m.err
}

func (m *errMockCli) ContainerList(_ context.Context, _ container.ListOptions) ([]types.Container, error) {
	m.containerListCount++
	return []types.Container{{}}, m.err
}

func (m *errMockCli) ImagePull(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
	m.imagePullCount++
	return io.NopCloser(&bytes.Buffer{}), m.err
}

func (m *errMockCli) Close() error {
	return nil
}

type mockDockerClient struct {
	client.APIClient
}

func newMockDockerClient(m *errMockCli) *core.DockerClient {
	return core.NewMockDockerClient(&mockDockerClient{
		APIClient: m,
	})
}

func (m *mockDockerClient) ClientVersion() string {
	return "mock-version"
}

func (m *mockDockerClient) Info(ctx context.Context) (system.Info, error) {
	return system.Info{}, nil
}

func (m *mockDockerClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return nil, nil
}

func (m *mockDockerClient) RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, nil
}

func (m *mockDockerClient) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return types.DiskUsage{}, nil
}

func (m *mockDockerClient) Ping(ctx context.Context) (types.Ping, error) {
	return types.Ping{}, nil
}

func (m *mockDockerClient) Close() error {
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
			errReturned: errdefs.NotFound(errors.New("not available")),
			shouldRetry: false,
		},
		{
			name:        "no retry when parameters are invalid",
			errReturned: errdefs.InvalidParameter(errors.New("invalid")),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access not authorized",
			errReturned: errdefs.Unauthorized(errors.New("not authorized")),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access is forbidden",
			errReturned: errdefs.Forbidden(errors.New("forbidden")),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.NotImplemented(errors.New("unknown method")),
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
			m := &errMockCli{err: tt.errReturned}

			// pass the mock client to the downstream API
			ctx := context.WithValue(context.Background(), ClientContextKey, newMockDockerClient(m))

			// give a chance to retry
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			_, _ = image.Build(ctx, &Request{})

			assert.Positive(t, m.imageBuildCount)
			assert.Equal(t, tt.shouldRetry, m.imageBuildCount > 1)
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
			errReturned: errdefs.InvalidParameter(errors.New("invalid")),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.NotImplemented(errors.New("unknown method")),
			shouldRetry: false,
		},
		{
			name:        "retry when not found",
			errReturned: errdefs.NotFound(errors.New("not there yet")),
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
			m := &errMockCli{err: tt.errReturned}

			ctx := context.WithValue(context.Background(), ClientContextKey, newMockDockerClient(m))

			// give a chance to retry
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			_, _ = waitContainerCreation(ctx, "someID")

			assert.Positive(t, m.containerListCount)
			assert.Equal(t, tt.shouldRetry, m.containerListCount > 1)
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
			errReturned: errdefs.NotFound(errors.New("not available")),
			shouldRetry: false,
		},
		{
			name:        "no retry when parameters are invalid",
			errReturned: errdefs.InvalidParameter(errors.New("invalid")),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access not authorized",
			errReturned: errdefs.Unauthorized(errors.New("not authorized")),
			shouldRetry: false,
		},
		{
			name:        "no retry when resource access is forbidden",
			errReturned: errdefs.Forbidden(errors.New("forbidden")),
			shouldRetry: false,
		},
		{
			name:        "no retry when not implemented by provider",
			errReturned: errdefs.NotImplemented(errors.New("unknown method")),
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
			m := &errMockCli{err: tt.errReturned, logger: tclog.NewTestLogger(t)}

			ctx := context.WithValue(context.Background(), ClientContextKey, newMockDockerClient(m))

			// give a chance to retry
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			_ = tcimage.Pull(ctx, "someTag", m.logger, image.PullOptions{})

			assert.Positive(t, m.imagePullCount)
			assert.Equal(t, tt.shouldRetry, m.imagePullCount > 1)
		})
	}
}
