package testcontainers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestContainerAttachedToNewNetwork(t *testing.T) {
	networkName := "new-network"
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
			NetworkAliases: map[string][]string{
				networkName: {
					"alias1", "alias2", "alias3",
				},
			},
		},
	}

	newNetwork, err := GenericNetwork(ctx, GenericNetworkRequest{
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer newNetwork.Remove(ctx)

	nginx, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}
	defer nginx.Terminate(ctx)

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
	if len(networkAlias) != 3 {
		t.Errorf("Expected network aliases %d. Got '%d'.", 3, len(networkAlias))
	}
	if networkAlias[0] != "alias1" || networkAlias[1] != "alias2" || networkAlias[2] != "alias3" {
		t.Errorf(
			"Expected network aliases '%s', '%s' and '%s'. Got '%s', '%s' and '%s'.",
			"alias1", "alias2", "alias3", networkAlias[0], networkAlias[1], networkAlias[2])
	}
}

func TestContainerWithHostNetworkOptions(t *testing.T) {
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:       "nginx",
			Privileged:  true,
			SkipReaper:  true,
			NetworkMode: "host",
			ExposedPorts: []string{
				"80/tcp",
			},
			WaitingFor: wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	defer nginxC.Terminate(ctx)

	//host, err := nginxC.Host(ctx)
	//if err != nil {
	//	t.Errorf("Expected host %s. Got '%d'.", host, err)
	//}
	//
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
		ContainerRequest: ContainerRequest{
			Image:       "nginx",
			SkipReaper:  true,
			NetworkMode: "host",
			Networks:    []string{"new-network"},
		},
		Started: true,
	}

	_, err := GenericContainer(ctx, gcr)
	if err != nil {
		// Error when NetworkMode = host and Network = []string{"bridge"}
		t.Logf("Can't use Network and NetworkMode together, %s", err)
	}
}

func TestContainerWithHostNetworkOptionsAndWaitStrategy(t *testing.T) {
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:       "nginx",
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	defer nginxC.Terminate(ctx)

	host, err := nginxC.Host(ctx)
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", host, err)
	}

	_, err = http.Get("http://" + host + ":80")
	if err != nil {
		t.Errorf("Expected OK response. Got '%d'.", err)
	}
}

func TestContainerWithHostNetworkAndEndpoint(t *testing.T) {
	nginxPort := "80/tcp"
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:       "nginx",
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort(nat.Port(nginxPort)),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	defer nginxC.Terminate(ctx)

	hostN, err := nginxC.Endpoint(ctx, "")
	if err != nil {
		t.Errorf("Expected host %s. Got '%d'.", hostN, err)
	}
	t.Log(hostN)

	_, err = http.Get("http://" + hostN)
	if err != nil {
		t.Errorf("Expected OK response. Got '%d'.", err)
	}
}

func TestContainerWithHostNetworkAndPortEndpoint(t *testing.T) {
	nginxPort := "80/tcp"
	ctx := context.Background()
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:       "nginx",
			SkipReaper:  true,
			NetworkMode: "host",
			WaitingFor:  wait.ForListeningPort(nat.Port(nginxPort)),
		},
		Started: true,
	}

	nginxC, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	defer nginxC.Terminate(ctx)

	origin, err := nginxC.PortEndpoint(ctx, nat.Port(nginxPort), "http")
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
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nginxA.Terminate(ctx)
	if nginxA.GetContainerID() == "" {
		t.Errorf("expected a containerID but we got an empty string.")
	}
}

func TestContainerStartsWithoutTheReaper(t *testing.T) {
	t.Skip("need to use the sessionID")
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	_, err = GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			SkipReaper: true,
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
	if len(resp) != 0 {
		t.Fatal("expected zero reaper running.")
	}
}

func TestContainerStartsWithTheReaper(t *testing.T) {
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	_, err = GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
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
	client, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
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

func TestContainerTerminationWithReaper(t *testing.T) {
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	containerID := nginxA.GetContainerID()
	resp, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.State.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ContainerInspect(ctx, containerID)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerTerminationWithoutReaper(t *testing.T) {
	ctx := context.Background()
	client, err := client.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	client.NegotiateAPIVersion(ctx)
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			SkipReaper: true,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	containerID := nginxA.GetContainerID()
	resp, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.State.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ContainerInspect(ctx, containerID)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestTwoContainersExposingTheSamePort(t *testing.T) {
	ctx := context.Background()
	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxA.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	nginxB, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			WaitingFor: wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxB.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	ipA, err := nginxA.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	portA, err := nginxA.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ipA, portA.Port()))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	ipB, err := nginxB.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	portB, err := nginxB.MappedPort(ctx, "80")
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:%s", ipB, portB.Port()))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreation(t *testing.T) {
	ctx := context.Background()

	nginxPort := "80/tcp"
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				nginxPort,
			},
			WaitingFor: wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()
	ip, err := nginxC.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := nginxC.MappedPort(ctx, "80")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
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

func TestContainerCreationWithName(t *testing.T) {
	ctx := context.Background()

	creationName := fmt.Sprintf("%s_%d", "test_container", time.Now().Unix())
	expectedName := "/" + creationName // inspect adds '/' in the beginning
	nginxPort := "80/tcp"
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				nginxPort,
			},
			WaitingFor: wait.ForListeningPort("80/tcp"),
			Name:       creationName,
			Networks:   []string{"bridge"},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()
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
	if network != "bridge" {
		t.Errorf("Expected network name '%s'. Got '%s'.", "bridge", network)
	}
	ip, err := nginxC.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := nginxC.MappedPort(ctx, "80")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func TestContainerCreationAndWaitForListeningPortLongEnough(t *testing.T) {
	ctx := context.Background()

	nginxPort := "80/tcp"
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				nginxPort,
			},
			WaitingFor: wait.ForListeningPort("80"), // default startupTimeout is 60s
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()
	origin, err := nginxC.PortEndpoint(ctx, nat.Port(nginxPort), "http")
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
	t.Skip("Wait needs to be fixed")
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				"80/tcp",
			},
			WaitingFor: wait.ForListeningPort("80").
				WithTimeout(1 * time.Second),
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
	t.Skip("Wait needs to be fixed")
	ctx := context.Background()

	nginxPort := "80/tcp"
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				nginxPort,
			},
			WaitingFor: wait.ForHTTP("/"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	origin, err := nginxC.PortEndpoint(ctx, nat.Port(nginxPort), "http")
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
	t.Skip("Wait needs to be fixed")
	ctx := context.Background()
	// delayed-nginx will wait 2s before opening port
	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "menedev/delayed-nginx:1.15.2",
			ExposedPorts: []string{
				"80/tcp",
			},
			WaitingFor: wait.ForHTTP("/").
				WithTimeout(1 * time.Second),
		},
		Started: true,
	})
	defer func() {
		err := nginxC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerCreationWaitsForLogContextTimeout(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("test context timeout").
			WithTimeout(1 * time.Second),
	}
	_, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err == nil {
		t.Error("Expected timeout")
	}
}

func TestContainerCreationWaitsForLog(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
	}
	mysqlC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		t.Log("terminating container")
		err := mysqlC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	host, _ := mysqlC.Host(ctx)
	p, _ := mysqlC.MappedPort(ctx, "3306/tcp")
	port := p.Int()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
		"root", "password", host, port, "database")

	db, err := sql.Open("mysql", connectionString)
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
	context := context.Background()
	t.Log("got context, creating container request")
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context: "./testresources",
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	t.Log("creating generic container request from container request")

	genContainerReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	t.Log("creating redis container")

	redisC, err := GenericContainer(context, genContainerReq)

	t.Log("created redis container")

	defer func() {
		t.Log("terminating redis container")
		err := redisC.Terminate(context)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("terminated redis container")
	}()

	t.Log("getting redis container endpoint")
	endpoint, err := redisC.Endpoint(context, "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("retrieved redis container endpoint")

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	t.Log("pinging redis")
	pong, err := client.Ping().Result()

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
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, genContainerReq)

	if err != nil {
		t.Fatal(err)
	}

	ep, err := c.Endpoint(ctx, "http")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(ep + "/env")

	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, ba, string(body))

	defer func() {
		t.Log("terminating container")
		err := c.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("terminated container")
	}()
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
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, genContainerReq)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("terminating container")
		err := c.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("terminated container")
	}()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	temp := strings.Split(string(out), "\n")

	if temp[0] != "Step 1/1 : FROM alpine" {
		t.Errorf("Expected stout firstline to be %s. Got '%s'.", "Step 1/2 : FROM alpine", temp[0])
	}

}

func TestContainerCreationWaitsForLogAndPortContextTimeout(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "mysql:latest",
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
		Image:        "nginx:1.17.6",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp"),
	}
	nginx, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		err := nginx.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("terminated nginx container")
	}()
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerCreationWaitingForHostPortWithoutBashThrowsAnError(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "nginx:1.17.6-alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp"),
	}
	nginx, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		err := nginx.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("terminated nginx container")
	}()
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerCreationWaitsForLogAndPort(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "mysql:latest",
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
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("terminating container")
		err := mysqlC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	host, _ := mysqlC.Host(ctx)
	p, _ := mysqlC.MappedPort(ctx, "3306/tcp")
	port := p.Int()
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
}

func TestCMD(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in a command to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := ContainerRequest{
		Image: "alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("command override!"),
		),
		Cmd: []string{"echo", "command override!"},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// defer not needed, but keeping it in for consistency
	defer c.Terminate(ctx)
}

func TestEntrypoint(t *testing.T) {
	/*
		echo a unique statement to ensure that we
		can pass in an entrypoint to the ContainerRequest
		and it will be run when we run the container
	*/

	ctx := context.Background()

	req := ContainerRequest{
		Image: "alpine",
		WaitingFor: wait.ForAll(
			wait.ForLog("entrypoint override!"),
		),
		Entrypoint: []string{"echo", "entrypoint override!"},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// defer not needed, but keeping it in for consistency
	defer c.Terminate(ctx)
}

func ExampleDockerProvider_CreateContainer() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer nginxC.Terminate(ctx)
}

func ExampleContainer_Host() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer nginxC.Terminate(ctx)
	ip, _ := nginxC.Host(ctx)
	println(ip)
}

func ExampleContainer_Start() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
	})
	defer nginxC.Terminate(ctx)
	nginxC.Start(ctx)
}

func ExampleContainer_MappedPort() {
	ctx := context.Background()
	req := ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer nginxC.Terminate(ctx)
	ip, _ := nginxC.Host(ctx)
	port, _ := nginxC.MappedPort(ctx, "80")
	http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
}

func TestContainerCreationWithBindAndVolume(t *testing.T) {
	absPath, err := filepath.Abs("./testresources/hello.sh")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()
	// Create a Docker client.
	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	dockerCli.NegotiateAPIVersion(ctx)
	// Create the volume.
	vol, err := dockerCli.VolumeCreate(ctx, volume.VolumeCreateBody{
		Driver: "local",
	})
	if err != nil {
		t.Fatal(err)
	}
	volumeName := vol.Name
	defer func() {
		ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cnl()
		err := dockerCli.VolumeRemove(ctx, volumeName, true)
		if err != nil {
			t.Fatal(err)
		}
	}()
	// Create the container that writes into the mounted volume.
	bashC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "bash",
			BindMounts:   map[string]string{absPath: "/hello.sh"},
			VolumeMounts: map[string]string{volumeName: "/data"},
			Cmd:          []string{"bash", "/hello.sh"},
			WaitingFor:   wait.ForLog("done"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cnl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cnl()
		err := bashC.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()
}

func TestContainerWithTmpFs(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: "busybox",
		Cmd:   []string{"sleep", "10"},
		Tmpfs: map[string]string{"/testtmpfs": "rw"},
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		t.Log("terminating container")
		err := container.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	path := "/testtmpfs/test.file"

	c, err := container.Exec(ctx, []string{"ls", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 1 {
		t.Fatalf("File %s should not have existed, expected return code 1, got %v", path, c)
	}

	c, err = container.Exec(ctx, []string{"touch", path})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should have been created successfully, expected return code 0, got %v", path, c)
	}

	c, err = container.Exec(ctx, []string{"ls", path})
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
			ContainerRequest: ContainerRequest{
				Image:      "postgres:12",
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

func TestContainerWithCustomHostname(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("some-nginx-%s-%d", t.Name(), rand.Int())
	hostname := fmt.Sprintf("my-nginx-%s-%d", t.Name(), rand.Int())
	req := ContainerRequest{
		Name:     name,
		Image:    "nginx",
		Hostname: hostname,
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		t.Log("terminating container")
		err := container.Terminate(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()
	if actualHostname := readHostname(t, container.GetContainerID()); actualHostname != hostname {
		t.Fatalf("expected hostname %s, got %s", hostname, actualHostname)
	}
}

// TODO: replace with proper API call
func readHostname(t *testing.T, containerId string) string {
	command := exec.Command("curl",
		"--silent",
		"--unix-socket",
		"/var/run/docker.sock",
		fmt.Sprintf("http://localhost/containers/%s/json", containerId))

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		t.Fatal(err)
	}
	config := data["Config"].(map[string]interface{})
	return config["Hostname"].(string)
}

func TestDockerContainerCopyFileToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
		},
		Started: true,
	})
	defer nginxC.Terminate(ctx)

	copiedFileName := "hello_copy.sh"
	nginxC.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)
	c, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}
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
			NetworkRequest: nr,
		})
		assert.Nil(t, err)
	}

	req := ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("80/tcp"),
			wait.ForLog("Configuration complete; ready for start up"),
		),
		Networks: networks,
	}

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	defer func() {
		t.Log("terminating container")
		err := nginxC.Terminate(ctx)
		assert.Nil(t, err)
	}()

	assert.Nil(t, err)
	containerId := nginxC.GetContainerID()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, err)
	cnt, err := cli.ContainerInspect(ctx, containerId)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cnt.NetworkSettings.Networks))
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[0]])
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[1]])
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
