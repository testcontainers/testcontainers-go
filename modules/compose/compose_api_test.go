package compose

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDockerComposeAPI(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NilError(t, compose.Up(ctx, Wait(true)), "compose.Up()")
}

func TestDockerComposeAPIStrategyForInvalidService(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Appending with _1 as given in the Java Test-Containers Example
		WaitForService("non-existent-srv-1", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.Assert(t, is.ErrorContains(err, ""), "Expected error to be thrown because service with wait strategy is not running")
	assert.Equal(t, "no container found for service name non-existent-srv-1", err.Error())

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIWithWaitLogStrategy(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
}

func TestDockerComposeAPIWithRunServices(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true), RunServices("api-nginx"))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	_, err = compose.ServiceContainer(context.Background(), "api-mysql")
	assert.Assert(t, is.ErrorContains(err, ""), "Make sure there is no mysql container")

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPI_TestcontainersLabelsArePresent(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))

	// all the services in the compose has the Testcontainers Labels
	for _, serviceName := range serviceNames {
		c, err := compose.ServiceContainer(context.Background(), serviceName)
		assert.NilError(t, err, "compose.ServiceContainer()")

		inspect, err := compose.dockerClient.ContainerInspect(ctx, c.GetContainerID())
		assert.NilError(t, err, "dockerClient.ContainerInspect()")

		for key, label := range testcontainers.GenericLabels() {
			assert.Check(t, is.Contains(inspect.Config.Labels, key), "Label %s is not present in container %s", key, c.GetContainerID())
			assert.Check(t, is.Equal(label, inspect.Config.Labels[key]), "Label %s value is not correct in container %s", key, c.GetContainerID())
		}
	}
}

func TestDockerComposeAPI_WithReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	// reaper is enabled, so we don't need to manually stop the containers: Ryuk will do it for us

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
}

func TestDockerComposeAPI_WithoutReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if !tcConfig.RyukDisabled {
		t.Skip("Ryuk is enabled, skipping test")
	}

	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")
	t.Cleanup(func() {
		// because reaper is disabled, we need to manually stop the containers
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
}

func TestDockerComposeAPIWithStopServices(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerComposeWith(
		WithStackFiles(path),
		WithLogger(testcontainers.TestLogger(t)))
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NilError(t, compose.Up(ctx, Wait(true)), "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))

	// close mysql container in purpose
	mysqlContainer, err := compose.ServiceContainer(context.Background(), "api-mysql")
	assert.NilError(t, err, "Get mysql container")

	stopTimeout := 10 * time.Second
	err = mysqlContainer.Stop(ctx, &stopTimeout)
	assert.NilError(t, err, "Stop mysql container")

	// check container status
	state, err := mysqlContainer.State(ctx)
	assert.NilError(t, err)
	assert.Check(t, !state.Running)
	assert.Check(t, is.Contains([]string{"exited", "removing"}, state.Status))
}

func TestDockerComposeAPIWithWaitForService(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIWithWaitHTTPStrategy(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIWithContainerName(t *testing.T) {
	path := RenderComposeWithName(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := RenderComposeWithoutExposedPorts(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-nginx", wait.ForLog("Configuration complete; ready for start up")).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIWithMultipleWaitStrategies(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
}

func TestDockerComposeAPIWithFailedStrategy(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx_1", wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Up(ctx, Wait(true))

	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.Assert(t, is.ErrorContains(err, ""), "Expected error to be thrown because of a wrong suplied wait strategy")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
}

func TestDockerComposeAPIComplex(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NilError(t, compose.Up(ctx, Wait(true)), "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
}

func TestDockerComposeAPIWithStackReader(t *testing.T) {
	identifier := testNameHash(t.Name())

	composeContent := `version: '3.7'
services:
  api-nginx:
    image: docker.io/nginx:stable-alpine
    environment:
      bar: ${bar}
      foo: ${foo}
`

	compose, err := NewDockerComposeWith(WithStackReaders(strings.NewReader(composeContent)), identifier)
	assert.NilError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"foo": "FOO",
			"bar": "BAR",
		}).
		Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))

	assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")

	// check files where removed
	f, err := os.Stat(compose.configs[0])
	assert.Assert(t, is.ErrorContains(err, ""), "File should be removed")
	assert.Assert(t, os.IsNotExist(err), "File should be removed")
	assert.Assert(t, is.Nil(f), "File should be removed")
}

func TestDockerComposeAPIWithStackReaderAndComposeFile(t *testing.T) {
	identifier := testNameHash(t.Name())
	simple, _ := RenderComposeSimple(t)
	composeContent := `version: '3.7'
services:
  api-postgres:
    image: docker.io/postgres:14
    environment:
      POSTGRES_PASSWORD: s3cr3t
`

	compose, err := NewDockerComposeWith(
		identifier,
		WithStackFiles(simple),
		WithStackReaders(strings.NewReader(composeContent)),
	)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 2))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-postgres"))

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier.String(), "api-nginx", present, absent)
}

func TestDockerComposeAPIWithEnvironment(t *testing.T) {
	identifier := testNameHash(t.Name())

	path, _ := RenderComposeSimple(t)

	compose, err := NewDockerComposeWith(WithStackFiles(path), identifier)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 1))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier.String(), "api-nginx", present, absent)
}

func TestDockerComposeAPIWithMultipleComposeFiles(t *testing.T) {
	simple, _ := RenderComposeSimple(t)
	composeFiles := ComposeStackFiles{
		simple,
		RenderComposePostgres(t),
		RenderComposeOverride(t),
	}

	identifier := testNameHash(t.Name())

	compose, err := NewDockerComposeWith(composeFiles, identifier)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	assert.Check(t, is.Len(serviceNames, 3))
	assert.Check(t, is.Contains(serviceNames, "api-nginx"))
	assert.Check(t, is.Contains(serviceNames, "api-mysql"))
	assert.Check(t, is.Contains(serviceNames, "api-postgres"))

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, identifier.String(), "api-nginx", present, absent)
}

func TestDockerComposeAPIWithVolume(t *testing.T) {
	path := RenderComposeWithVolume(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")
}

func TestDockerComposeAPIWithRecreate(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, WithRecreate(api.RecreateNever), WithRecreateDependencies(api.RecreateNever), Wait(true))
	assert.NilError(t, err, "compose.Up()")
}

func TestDockerComposeAPIVolumesDeletedOnDown(t *testing.T) {
	path := RenderComposeWithVolume(t)
	identifier := uuid.New().String()
	stackFiles := WithStackFiles(path)
	compose, err := NewDockerComposeWith(stackFiles, StackIdentifier(identifier))
	assert.NilError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	assert.NilError(t, err, "compose.Up()")

	err = compose.Down(context.Background(), RemoveOrphans(true), RemoveVolumes(true), RemoveImagesLocal)
	assert.NilError(t, err, "compose.Down()")

	volumeListFilters := filters.NewArgs()
	// the "mydata" identifier comes from the "testdata/docker-compose-volume.yml" file
	volumeListFilters.Add("name", fmt.Sprintf("%s_mydata", identifier))
	volumeList, err := compose.dockerClient.VolumeList(ctx, volume.ListOptions{Filters: volumeListFilters})
	assert.NilError(t, err, "compose.dockerClient.VolumeList()")

	assert.Check(t, is.Len(volumeList.Volumes, 0), "Volumes are not cleaned up")
}

func TestDockerComposeAPIWithBuild(t *testing.T) {
	t.Skip("Skipping test because of the opentelemetry dependencies issue. See https://github.com/open-telemetry/opentelemetry-go/issues/4476#issuecomment-1840547010")

	path := RenderComposeWithBuild(t)
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-echo", wait.ForHTTP("/env").WithPort("8080/tcp")).
		Up(ctx, Wait(true))

	assert.NilError(t, err, "compose.Up()")
}

func TestDockerComposeApiWithWaitForShortLifespanService(t *testing.T) {
	path := filepath.Join(testdataPackage, "docker-compose-short-lifespan.yml")
	compose, err := NewDockerCompose(path)
	assert.NilError(t, err, "NewDockerCompose()")

	t.Cleanup(func() {
		assert.NilError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Assumption: tzatziki service wait logic will run before falafel, so that falafel service will exit before
		WaitForService("tzatziki", wait.ForExit().WithExitTimeout(10*time.Second)).
		WaitForService("falafel", wait.ForExit().WithExitTimeout(10*time.Second)).
		Up(ctx)

	assert.NilError(t, err, "compose.Up()")

	services := compose.Services()

	assert.Check(t, is.Len(services, 2))
	assert.Check(t, is.Contains(services, "falafel"))
	assert.Check(t, is.Contains(services, "tzatziki"))
}

func testNameHash(name string) StackIdentifier {
	return StackIdentifier(fmt.Sprintf("%x", fnv.New32a().Sum([]byte(name))))
}
