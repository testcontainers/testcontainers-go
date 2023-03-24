package compose

import (
	"context"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	simpleCompose     = "docker-compose-simple.yml"
	complexCompose    = "docker-compose-complex.yml"
	composeWithVolume = "docker-compose-volume.yml"
	testdataPackage   = "testdata"
)

func TestDockerComposeAPI(t *testing.T) {
	path := filepath.Join(testdataPackage, simpleCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NoError(t, compose.Up(ctx, Wait(true)), "compose.Up()")
}

func TestDockerComposeAPIStrategyForInvalidService(t *testing.T) {
	path := filepath.Join(testdataPackage, simpleCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	assert.Equal(t, err.Error(), "no container found for service name mysql-1")

	serviceNames := compose.Services()

	assert.Equal(t, 1, len(serviceNames))
	assert.Contains(t, serviceNames, "nginx")
}

func TestDockerComposeAPIWithWaitLogStrategy(t *testing.T) {
	path := filepath.Join(testdataPackage, complexCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, complexCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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

func TestDockerComposeAPIWithStopServices(t *testing.T) {
	path := filepath.Join(testdataPackage, complexCompose)
	compose, err := NewDockerComposeWith(
		WithStackFiles(path),
		WithLogger(testcontainers.TestLogger(t)))
	assert.NoError(t, err, "NewDockerCompose()")

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

	// close mysql container in purpose
	mysqlContainer, err := compose.ServiceContainer(context.Background(), "mysql")
	assert.NoError(t, err, "Get mysql container")

	stopTimeout := 10 * time.Second
	err = mysqlContainer.Stop(ctx, &stopTimeout)
	assert.NoError(t, err, "Stop mysql container")

	// check container status
	state, err := mysqlContainer.State(ctx)
	assert.NoError(t, err)
	assert.False(t, state.Running)
	assert.Equal(t, "exited", state.Status)
}

func TestDockerComposeAPIWithWaitForService(t *testing.T) {
	path := filepath.Join(testdataPackage, simpleCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, simpleCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, "docker-compose-container-name.yml")
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, "docker-compose-no-exposed-ports.yml")
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, complexCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, simpleCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	path := filepath.Join(testdataPackage, complexCompose)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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
	identifier := testNameHash(t.Name())

	path := filepath.Join(testdataPackage, simpleCompose)

	compose, err := NewDockerComposeWith(WithStackFiles(path), identifier)
	assert.NoError(t, err, "NewDockerCompose()")

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
	assertContainerEnvironmentVariables(t, identifier.String(), "nginx", present, absent)
}

func TestDockerComposeAPIWithMultipleComposeFiles(t *testing.T) {
	composeFiles := ComposeStackFiles{
		filepath.Join(testdataPackage, simpleCompose),
		filepath.Join(testdataPackage, "docker-compose-postgres.yml"),
		filepath.Join(testdataPackage, "docker-compose-override.yml"),
	}

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeWith(composeFiles, identifier)
	assert.NoError(t, err, "NewDockerCompose()")

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
	assertContainerEnvironmentVariables(t, identifier.String(), "nginx", present, absent)
}

func TestDockerComposeAPIWithVolume(t *testing.T) {
	path := filepath.Join(testdataPackage, composeWithVolume)
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	assert.NoError(t, err, "compose.Up()")
}

func TestDockerComposeAPIVolumesDeletedOnDown(t *testing.T) {
	path := filepath.Join(testdataPackage, composeWithVolume)
	identifier := uuid.New().String()
	stackFiles := WithStackFiles(path)
	compose, err := NewDockerComposeWith(stackFiles, StackIdentifier(identifier))
	assert.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	assert.NoError(t, err, "compose.Up()")

	err = compose.Down(context.Background(), RemoveOrphans(true), RemoveVolumes(true), RemoveImagesLocal)
	assert.NoError(t, err, "compose.Down()")

	volumeListFilters := filters.NewArgs()
	// the "mydata" identifier comes from the "testdata/docker-compose-volume.yml" file
	volumeListFilters.Add("name", fmt.Sprintf("%s_mydata", identifier))
	volumeList, err := compose.dockerClient.VolumeList(ctx, volumeListFilters)
	assert.NoError(t, err, "compose.dockerClient.VolumeList()")

	assert.Equal(t, 0, len(volumeList.Volumes), "Volumes are not cleaned up")
}

func TestDockerComposeAPIWithBuild(t *testing.T) {
	path := filepath.Join(testdataPackage, "docker-compose-build.yml")
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

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

func TestDockerComposeApiWithWaitForShortLifespanService(t *testing.T) {
	path := filepath.Join(testdataPackage, "docker-compose-short-lifespan.yml")
	compose, err := NewDockerCompose(path)
	assert.NoError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		//Assumption: tzatziki service wait logic will run before falafel, so that falafel service will exit before
		WaitForService("tzatziki", wait.ForExit().WithExitTimeout(10*time.Second)).
		WaitForService("falafel", wait.ForExit().WithExitTimeout(10*time.Second)).
		Up(ctx)

	assert.NoError(t, err, "compose.Up()")

	services := compose.Services()

	assert.Equal(t, 2, len(services))
	assert.Contains(t, services, "falafel")
	assert.Contains(t, services, "tzatziki")
}

func testNameHash(name string) StackIdentifier {
	return StackIdentifier(fmt.Sprintf("%x", fnv.New32a().Sum([]byte(name))))
}
