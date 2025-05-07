package compose

import (
	"context"
	"encoding/hex"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDockerComposeAPI(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	for _, service := range compose.Services() {
		container, err := compose.ServiceContainer(context.Background(), service)
		require.NoError(t, err, "compose.ServiceContainer()")
		require.True(t, container.IsRunning())
	}
}

func TestDockerComposeAPIStrategyForInvalidService(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Appending with _1 as given in the Java Test-Containers Example
		WaitForService("non-existent-srv-1", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.EqualError(t, err, "wait for services: no container found for service name non-existent-srv-1")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithWaitLogStrategy(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")
}

func TestDockerComposeAPIWithRunServices(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true), RunServices("api-nginx"))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	_, err = compose.ServiceContainer(context.Background(), "api-mysql")
	require.Error(t, err, "Make sure there is no mysql container")

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithProfiles(t *testing.T) {
	path := RenderComposeProfiles(t)

	testcases := map[string]struct {
		withProfiles []string
		wantServices []string
	}{
		"nil profile": {
			withProfiles: nil,
			wantServices: []string{"starts-always"},
		},
		"no profiles": {
			withProfiles: []string{},
			wantServices: []string{"starts-always"},
		},
		"dev profile": {
			withProfiles: []string{"dev"},
			wantServices: []string{"starts-always", "only-dev", "dev-or-test"},
		},
		"test profile": {
			withProfiles: []string{"test"},
			wantServices: []string{"starts-always", "dev-or-test"},
		},
		"wildcard profile": {
			withProfiles: []string{"*"},
			wantServices: []string{"starts-always", "only-dev", "dev-or-test", "only-prod"},
		},
		"undefined profile": {
			withProfiles: []string{"undefined-profile"},
			wantServices: []string{"starts-always"},
		},
	}

	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {
			var compose ComposeStack
			compose, err := NewDockerComposeWith(WithStackFiles(path), WithProfiles(test.withProfiles...))
			require.NoError(t, err, "NewDockerCompose()")

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			for _, service := range test.wantServices {
				compose = compose.WaitForService(service, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second))
			}
			err = compose.Up(ctx, Wait(true))
			cleanup(t, compose)
			require.NoError(t, err, "compose.Up()")

			assert.ElementsMatch(t, test.wantServices, compose.Services())
		})
	}
}

func TestDockerComposeAPI_TestcontainersLabelsArePresent(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")

	// all the services in the compose have the Testcontainers Labels
	for _, serviceName := range serviceNames {
		c, err := compose.ServiceContainer(context.Background(), serviceName)
		require.NoError(t, err, "compose.ServiceContainer()")

		inspect, err := compose.dockerClient.ContainerInspect(ctx, c.GetContainerID())
		require.NoError(t, err, "dockerClient.ContainerInspect()")

		for key, label := range testcontainers.GenericLabels() {
			assert.Contains(t, inspect.Config.Labels, key, "Label %s is not present in container %s", key, c.GetContainerID())
			assert.Equal(t, label, inspect.Config.Labels[key], "Label %s value is not correct in container %s", key, c.GetContainerID())
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
	require.NoError(t, err, "NewDockerCompose()")

	// reaper is enabled, so we don't need to manually stop the containers: Ryuk will do it for us

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")
}

func TestDockerComposeAPI_WithoutReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if !tcConfig.RyukDisabled {
		t.Skip("Ryuk is enabled, skipping test")
	}

	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")
}

func TestDockerComposeAPIWithStopServices(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerComposeWith(
		WithStackFiles(path),
		WithLogger(log.TestLogger(t)))
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")

	// close mysql container in purpose
	mysqlContainer, err := compose.ServiceContainer(context.Background(), "api-mysql")
	require.NoError(t, err, "Get mysql container")

	stopTimeout := 10 * time.Second
	err = mysqlContainer.Stop(ctx, &stopTimeout)
	require.NoError(t, err, "Stop mysql container")

	// check container status
	state, err := mysqlContainer.State(ctx)
	require.NoError(t, err)
	assert.False(t, state.Running)
	assert.Contains(t, []string{"exited", "removing"}, state.Status)
}

func TestDockerComposeAPIWithWaitForService(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithWaitHTTPStrategy(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithContainerName(t *testing.T) {
	path := RenderComposeWithName(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := RenderComposeWithoutExposedPorts(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-nginx", wait.ForLog("Configuration complete; ready for start up")).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIWithMultipleWaitStrategies(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-mysql", wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WaitForService("api-nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")
}

func TestDockerComposeAPIWithFailedStrategy(t *testing.T) {
	path, _ := RenderComposeSimple(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("api-nginx_1", wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	require.Error(t, err, "Expected error to be thrown because of a wrong supplied wait strategy")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")
}

func TestDockerComposeAPIComplex(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	require.Contains(t, serviceNames, "api-nginx")
	require.Contains(t, serviceNames, "api-mysql")
}

func TestDockerComposeAPIWithStackReader(t *testing.T) {
	identifier := testNameHash(t.Name())

	composeContent := `
services:
  api-nginx:
    image: nginx:stable-alpine
    environment:
      bar: ${bar}
      foo: ${foo}
`

	compose, err := NewDockerComposeWith(WithStackReaders(strings.NewReader(composeContent)), identifier)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"foo": "FOO",
			"bar": "BAR",
		}).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")

	require.NoError(t, compose.Down(context.Background(), RemoveOrphans(true), RemoveVolumes(true), RemoveImagesLocal), "compose.Down()")

	// check files where removed
	f, err := os.Stat(compose.configs[0])
	require.Error(t, err, "File should be removed")
	require.True(t, os.IsNotExist(err), "File should be removed")
	require.Nil(t, f, "File should be removed")
}

func TestDockerComposeAPIWithStackReaderAndComposeFile(t *testing.T) {
	identifier := testNameHash(t.Name())
	simple, _ := RenderComposeSimple(t)
	composeContent := `
services:
  api-postgres:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: s3cr3t
`

	compose, err := NewDockerComposeWith(
		identifier,
		WithStackFiles(simple),
		WithStackReaders(strings.NewReader(composeContent)),
	)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 2)
	assert.Contains(t, serviceNames, "api-nginx")
	assert.Contains(t, serviceNames, "api-postgres")

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
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 1)
	assert.Contains(t, serviceNames, "api-nginx")

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
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	serviceNames := compose.Services()

	require.Len(t, serviceNames, 3)
	assert.Contains(t, serviceNames, "api-nginx")
	assert.Contains(t, serviceNames, "api-mysql")
	assert.Contains(t, serviceNames, "api-postgres")

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
	require.NoError(t, err, "NewDockerCompose()")

	cleanup(t, compose)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	require.NoError(t, err, "compose.Up()")
}

func TestDockerComposeAPIWithRecreate(t *testing.T) {
	path, _ := RenderComposeComplex(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, WithRecreate(api.RecreateNever), WithRecreateDependencies(api.RecreateNever), Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")
}

func TestDockerComposeAPIVolumesDeletedOnDown(t *testing.T) {
	path := RenderComposeWithVolume(t)
	identifier := uuid.New().String()
	stackFiles := WithStackFiles(path)
	compose, err := NewDockerComposeWith(stackFiles, StackIdentifier(identifier))
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	err = compose.Down(context.Background(), RemoveOrphans(true), RemoveVolumes(true), RemoveImagesLocal)
	require.NoError(t, err, "compose.Down()")

	volumeListFilters := filters.NewArgs()
	// the "mydata" identifier comes from the "testdata/docker-compose-volume.yml" file
	volumeListFilters.Add("name", identifier+"_mydata")
	volumeList, err := compose.dockerClient.VolumeList(ctx, volume.ListOptions{Filters: volumeListFilters})
	require.NoError(t, err, "compose.dockerClient.VolumeList()")

	require.Empty(t, volumeList.Volumes, "Volumes are not cleaned up")
}

func TestDockerComposeAPIWithBuild(t *testing.T) {
	t.Skip("Skipping test because of the opentelemetry dependencies issue. See https://github.com/open-telemetry/opentelemetry-go/issues/4476#issuecomment-1840547010")

	path := RenderComposeWithBuild(t)
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		WaitForService("api-echo", wait.ForHTTP("/env").WithPort("8080/tcp")).
		Up(ctx, Wait(true))
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")
}

func TestDockerComposeApiWithWaitForShortLifespanService(t *testing.T) {
	path := filepath.Join(testdataPackage, "docker-compose-short-lifespan.yml")
	compose, err := NewDockerCompose(path)
	require.NoError(t, err, "NewDockerCompose()")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.
		// Assumption: tzatziki service wait logic will run before falafel, so that falafel service will exit before
		WaitForService("tzatziki", wait.ForExit().WithExitTimeout(10*time.Second)).
		WaitForService("falafel", wait.ForExit().WithExitTimeout(10*time.Second)).
		Up(ctx)
	cleanup(t, compose)
	require.NoError(t, err, "compose.Up()")

	services := compose.Services()

	require.Len(t, services, 2)
	assert.Contains(t, services, "falafel")
	assert.Contains(t, services, "tzatziki")
}

func testNameHash(name string) StackIdentifier {
	return StackIdentifier(hex.EncodeToString(fnv.New32a().Sum([]byte(name))))
}

// cleanup is a helper function that schedules the compose stack to be stopped when the test ends.
func cleanup(t *testing.T, compose ComposeStack) {
	t.Helper()
	t.Cleanup(func() {
		require.NoError(t, compose.Down(
			context.Background(),
			RemoveOrphans(true),
			RemoveVolumes(true),
			RemoveImagesLocal,
		), "compose.Down()")
	})
}
