package testcontainers

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"database/sql"
	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
				networkName: []string{
					"alias1", "alias2", "alias3",
				},
			},
		},
	}

	provider, err := gcr.ProviderType.GetProvider()

	provider.CreateNetwork(ctx, NetworkRequest{
		Name:           networkName,
		CheckDuplicate: true,
	})
	defer provider.RemoveNetwork(ctx, NetworkRequest{
		Name: networkName,
	})

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
			Name:     creationName,
			Networks: []string{"bridge"},
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
	t.Skip("Wait needs to be fixed")
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
			WaitingFor: wait.ForListeningPort("80").WithStartupTimeout(1 * time.Second),
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

func TestContainerRespondsWithHttp404ForNonExistingPage(t *testing.T) {
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
			WaitingFor: wait.ForHTTP("/nonExistingPage").WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusNotFound
			}),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	rC, err := RunContainer(ctx, "nginx", RequestContainer{
		ExportedPort: []string{
			nginxPort,
		},
		WaitingFor: wait.ForHTTP("/nonExistingPage").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNotFound
		}),
	})
	if rC != nil {
		t.Fatal(rC)
	}
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
	resp, err := http.Get(origin + "/nonExistingPage")
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d. Got %d.", http.StatusNotFound, resp.StatusCode)
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
			WaitingFor: wait.ForHTTP("/").WithStartupTimeout(1 * time.Second),
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
		WaitingFor: wait.ForLog("test context timeout").WithStartupTimeout(1 * time.Second),
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
