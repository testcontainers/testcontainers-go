package registry_test

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRegistry_unauthenticated(t *testing.T) {
	container, err := registry.Run(context.Background(), "registry:2.8.3")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	httpAddress, err := container.Address(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(httpAddress + "/v2/_catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, but got %d", resp.StatusCode)
	}
}

func TestRunContainer_authenticated(t *testing.T) {
	registryContainer, err := registry.Run(
		context.Background(),
		"registry:2.8.3",
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "data")),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	t.Cleanup(func() {
		if err := registryContainer.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// httpAddress {
	httpAddress, err := registryContainer.Address(context.Background())
	// }
	if err != nil {
		t.Fatal(err)
	}

	registryPort, err := registryContainer.MappedPort(context.Background(), "5000/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)
	}
	strPort := registryPort.Port()

	t.Run("HTTP connection without basic auth fails", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest("GET", httpAddress+"/v2/_catalog", nil)
		if err != nil {
			tt.Fatal(err)
		}

		resp, err := httpCli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status code 401, but got %d", resp.StatusCode)
		}
	})

	t.Run("HTTP connection with incorrect basic auth fails", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest("GET", httpAddress+"/v2/_catalog", nil)
		if err != nil {
			tt.Fatal(err)
		}

		req.SetBasicAuth("foo", "bar")

		resp, err := httpCli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status code 401, but got %d", resp.StatusCode)
		}
	})

	t.Run("HTTP connection with basic auth succeeds", func(tt *testing.T) {
		httpCli := http.Client{}
		req, err := http.NewRequest("GET", httpAddress+"/v2/_catalog", nil)
		if err != nil {
			tt.Fatal(err)
		}

		req.SetBasicAuth("testuser", "testpassword")

		resp, err := httpCli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code 200, but got %d", resp.StatusCode)
		}
	})

	t.Run("build images with wrong credentials fails", func(tt *testing.T) {
		// Zm9vOmJhcg== is base64 for foo:bar
		tt.Setenv("DOCKER_AUTH_CONFIG", `{
			"auths": {
				"localhost:`+strPort+`": { "username": "foo", "password": "bar", "auth": "Zm9vOmJhcg==" }
			},
			"credsStore": "desktop"
		}`)

		redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: filepath.Join("testdata", "redis"),
					BuildArgs: map[string]*string{
						"REGISTRY_PORT": &strPort,
					},
					PrintBuildLog: true,
				},
				AlwaysPullImage: true, // make sure the authentication takes place
				ExposedPorts:    []string{"6379/tcp"},
				WaitingFor:      wait.ForLog("Ready to accept connections"),
			},
			Started: true,
		})
		if err == nil {
			tt.Fatalf("expected to fail to start container, but it did not")
		}
		if redisC != nil {
			tt.Fatal("redis container should not be running")
			tt.Cleanup(func() {
				if err := redisC.Terminate(context.Background()); err != nil {
					tt.Fatalf("failed to terminate container: %s", err)
				}
			})
		}

		if !strings.Contains(err.Error(), "unauthorized: authentication required") {
			tt.Fatalf("expected error to be 'unauthorized: authentication required' but got '%s'", err.Error())
		}
	})

	t.Run("build image with valid credentials", func(tt *testing.T) {
		// dGVzdHVzZXI6dGVzdHBhc3N3b3Jk is base64 for testuser:testpassword
		tt.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
			"localhost:`+strPort+`": { "username": "testuser", "password": "testpassword", "auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" }
		},
		"credsStore": "desktop"
	}`)

		// build a custom redis image from the private registry,
		// using RegistryName of the container as the registry.
		// The container should start because the authentication
		// is correct.

		redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context: filepath.Join("testdata", "redis"),
					BuildArgs: map[string]*string{
						"REGISTRY_PORT": &strPort,
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
			tt.Fatalf("failed to start container: %s", err)
		}

		tt.Cleanup(func() {
			if err := redisC.Terminate(context.Background()); err != nil {
				tt.Fatalf("failed to terminate container: %s", err)
			}
		})

		state, err := redisC.State(context.Background())
		if err != nil {
			tt.Fatalf("failed to get redis container state: %s", err) // nolint:gocritic
		}

		if !state.Running {
			tt.Fatalf("expected redis container to be running, but it is not")
		}
	})
}

func TestRunContainer_authenticated_withCredentials(t *testing.T) {
	// htpasswdString {
	registryContainer, err := registry.Run(
		context.Background(),
		"registry:2.8.3",
		registry.WithHtpasswd("testuser:$2y$05$tTymaYlWwJOqie.bcSUUN.I.kxmo1m5TLzYQ4/ejJ46UMXGtq78EO"),
	)
	// }
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	t.Cleanup(func() {
		if err := registryContainer.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	httpAddress, err := registryContainer.Address(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	httpCli := http.Client{}
	req, err := http.NewRequest("GET", httpAddress+"/v2/_catalog", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth("testuser", "testpassword")

	resp, err := httpCli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, but got %d", resp.StatusCode)
	}
}

func TestRunContainer_wrongData(t *testing.T) {
	registryContainer, err := registry.Run(
		context.Background(),
		"registry:2.8.3",
		registry.WithHtpasswdFile(filepath.Join("testdata", "auth", "htpasswd")),
		registry.WithData(filepath.Join("testdata", "wrongdata")),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	t.Cleanup(func() {
		if err := registryContainer.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	registryPort, err := registryContainer.MappedPort(context.Background(), "5000/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)
	}
	strPort := registryPort.Port()

	// dGVzdHVzZXI6dGVzdHBhc3N3b3Jk is base64 for testuser:testpassword
	t.Setenv("DOCKER_AUTH_CONFIG", `{
		"auths": {
			"localhost:`+strPort+`": { "username": "testuser", "password": "testpassword", "auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" }
		},
		"credsStore": "desktop"
	}`)

	// build a custom redis image from the private registry,
	// using RegistryName of the container as the registry.
	// The container won't be able to start because the data
	// directory is wrong.

	redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: filepath.Join("testdata", "redis"),
				BuildArgs: map[string]*string{
					"REGISTRY_PORT": &strPort,
				},
				PrintBuildLog: true,
			},
			AlwaysPullImage: true, // make sure the authentication takes place
			ExposedPorts:    []string{"6379/tcp"},
			WaitingFor:      wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err == nil {
		t.Fatalf("expected to fail to start container, but it did not")
	}
	if redisC != nil {
		t.Fatal("redis container should not be running")
		t.Cleanup(func() {
			if err := redisC.Terminate(context.Background()); err != nil {
				t.Fatalf("failed to terminate container: %s", err)
			}
		})
	}

	if !strings.Contains(err.Error(), "manifest unknown") {
		t.Fatalf("expected error to be 'manifest unknown' but got '%s'", err.Error())
	}
}
