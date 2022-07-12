package testcontainers

import (
	"context"
	"fmt"
	"hash/fnv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDockerComposeAPI(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NoError(t, compose.Up(ctx, Wait(true)), "compose.Up()")
}

func TestDockerComposeAPIStrategyForInvalidService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Appending with _1 as given in the Java Test-Containers Example
		WaitForService("mysql-1", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.Error(t, err, "Expected error to be thrown because service with wait strategy is not running")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithWaitLogStrategy(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeAPIWithRunServices(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true), RunServices("nginx"))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	_, err = compose.ServiceContainer(context.Background(), "mysql")
	assert.Error(t, err, "Make sure there is no mysql container")

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithWaitForService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithWaitHTTPStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithContainerName(t *testing.T) {
	path := "./testresources/docker-compose-container-name.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := "./testresources/docker-compose-no-exposed-ports.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("nginx", wait.ForLog("Configuration complete; ready for start up")).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithMultipleWaitStrategies(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeAPIWithFailedStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx_1", wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Up(ctx, Wait(true))

	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.Error(t, err, "Expected error to be thrown because of a wrong suplied wait strategy")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIComplex(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NoError(t, compose.Up(ctx, Wait(true)), "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeAPIWithEnvironment(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier, "nginx", present, absent)
}

func TestDockerComposeAPIWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		"testresources/docker-compose-simple.yml",
		"testresources/docker-compose-postgres.yml",
		"testresources/docker-compose-override.yml",
	}

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI(composeFiles, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx, Wait(true))
	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Equal(t, 3, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
	assert.Contains(t, serviceNames, "postgres")

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier, "nginx", present, absent)
}

func TestDockerComposeAPIWithVolume(t *testing.T) {
	path := "./testresources/docker-compose-volume.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	assert.NoError(t, err, "compose.Up()")
}

func TestDockerComposeAPIWithBuild(t *testing.T) {
	path := "./testresources/docker-compose-build.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeAPI([]string{path}, identifier)
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("echo", wait.ForHTTP("/env").WithPort("8080/tcp")).
		Up(ctx, Wait(true))

	assert.NoError(t, err, "compose.Up()")
}

func testNameHash(name string) string {
	return fmt.Sprintf("%x", fnv.New32a().Sum([]byte(name)))
}
