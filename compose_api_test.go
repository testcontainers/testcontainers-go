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

func TestDockerComposeApi(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NoError(t, compose.Up(ctx), "compose.Up()")
}

func TestDockerComposeApiStrategyForInvalidService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql-1", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx)

	assert.Error(t, err, "Expected error to be thrown because service with wait strategy is not running")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiWithWaitLogStrategy(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeApiWithWaitForService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiWithWaitHTTPStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginx", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiWithContainerName(t *testing.T) {
	path := "./testresources/docker-compose-container-name.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginx", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := "./testresources/docker-compose-no-exposed-ports.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithExposedService("nginx", 9080, wait.ForLog("Configuration complete; ready for start up")).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiWithMultipleWaitStrategies(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithExposedService("mysql", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WithExposedService("nginx", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeApiWithFailedStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Up(ctx)

	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.Error(t, err, "Expected error to be thrown because of a wrong suplied wait strategy")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeApiComplex(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NoError(t, compose.Up(ctx), "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 2, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
	assert.Contains(t, serviceNames, "mysql")
}

func TestDockerComposeApiWithEnvironment(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier, "nginx", present, absent)
}

func TestDockerComposeApiWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		"testresources/docker-compose-simple.yml",
		"testresources/docker-compose-postgres.yml",
		"testresources/docker-compose-override.yml",
	}

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi(composeFiles, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx)
	assert.NoError(t, err, "compose.Up()")

	serviceNames := compose.ServiceNames()

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

func TestDockerComposeApiWithVolume(t *testing.T) {
	path := "./testresources/docker-compose-volume.yml"

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeApi([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.NoError(t, err, "NewDockerComposeApi()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background()), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx)
	assert.NoError(t, err, "compose.Up()")
}

func testNameHash(name string) string {
	return fmt.Sprintf("%x", fnv.New32a().Sum([]byte(name)))
}
