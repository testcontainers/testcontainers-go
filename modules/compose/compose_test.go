package compose

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLocalDockerCompose(t *testing.T) {
	path, _ := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)
}

func TestLocalDockerComposeStrategyForInvalidService(t *testing.T) {
	path, ports := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService(compose.Format("non-existent-srv", "1"), ports[0], wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	require.Error(t, err.Error, "Expected error to be thrown because service with wait strategy is not running")

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeWithWaitLogStrategy(t *testing.T) {
	path, _ := RenderComposeComplexForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService(compose.Format("local-mysql", "1"), 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 2)
	assert.Contains(t, compose.Services, "local-nginx")
	assert.Contains(t, compose.Services, "local-mysql")
}

func TestLocalDockerComposeWithWaitForService(t *testing.T) {
	path, _ := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService(compose.Format("local-nginx", "1"), wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeWithWaitForShortLifespanService(t *testing.T) {
	path := filepath.Join(testdataPackage, "docker-compose-short-lifespan.yml")

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Assumption: tzatziki service wait logic will run before falafel, so that falafel service will exit before
		WaitForService(compose.Format("tzatziki", "1"), wait.ForExit().WithExitTimeout(10*time.Second)).
		WaitForService(compose.Format("falafel", "1"), wait.ForExit().WithExitTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 2)
	assert.Contains(t, compose.Services, "falafel")
	assert.Contains(t, compose.Services, "tzatziki")
}

func TestLocalDockerComposeWithWaitHTTPStrategy(t *testing.T) {
	path, ports := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService(compose.Format("local-nginx", "1"), ports[0], wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeWithContainerName(t *testing.T) {
	path := RenderComposeWithNameForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("local-nginxy", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := RenderComposeWithoutExposedPortsForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService(compose.Format("local-nginx", "1"), 9080, wait.ForLog("Configuration complete; ready for start up")).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeWithMultipleWaitStrategies(t *testing.T) {
	path, _ := RenderComposeComplexForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService(compose.Format("local-mysql", "1"), 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WithExposedService(compose.Format("local-nginx", "1"), 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 2)
	assert.Contains(t, compose.Services, "local-nginx")
	assert.Contains(t, compose.Services, "local-mysql")
}

func TestLocalDockerComposeWithFailedStrategy(t *testing.T) {
	path, ports := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("local-nginx_1", ports[0], wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Invoke()
	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	require.Error(t, err.Error, "Expected error to be thrown because of a wrong supplied wait strategy")

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")
}

func TestLocalDockerComposeComplex(t *testing.T) {
	path, _ := RenderComposeComplexForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 2)
	assert.Contains(t, compose.Services, "local-nginx")
	assert.Contains(t, compose.Services, "local-mysql")
}

func TestLocalDockerComposeWithEnvironment(t *testing.T) {
	path, _ := RenderComposeSimpleForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 1)
	assert.Contains(t, compose.Services, "local-nginx")

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier, "local-nginx", present, absent)
}

func TestLocalDockerComposeWithMultipleComposeFiles(t *testing.T) {
	simple, _ := RenderComposeSimpleForLocal(t)
	composeFiles := []string{
		simple,
		RenderComposePostgresForLocal(t),
		RenderComposeOverrideForLocal(t),
	}

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose(composeFiles, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Invoke()
	checkIfError(t, err)

	require.Len(t, compose.Services, 3)
	assert.Contains(t, compose.Services, "local-nginx")
	assert.Contains(t, compose.Services, "local-mysql")
	assert.Contains(t, compose.Services, "local-postgres")

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier, "local-nginx", present, absent)
}

func TestLocalDockerComposeWithVolume(t *testing.T) {
	path := RenderComposeWithVolumeForLocal(t)

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(log.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
		assertVolumeDoesNotExist(t, compose.Format(identifier, "mydata"))
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)
}

func assertVolumeDoesNotExist(tb testing.TB, volumeName string) {
	tb.Helper()
	containerClient, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoErrorf(tb, err, "Failed to get provider")

	volumeList, err := containerClient.VolumeList(context.Background(), volume.ListOptions{Filters: filters.NewArgs(filters.Arg("name", volumeName))})
	require.NoErrorf(tb, err, "Failed to list volumes")

	if len(volumeList.Warnings) > 0 {
		tb.Logf("Volume list warnings: %v", volumeList.Warnings)
	}

	require.Emptyf(tb, volumeList.Volumes, "Volume list is not empty")
}

func assertContainerEnvironmentVariables(
	tb testing.TB,
	composeIdentifier, serviceName string,
	present map[string]string,
	absent map[string]string,
) {
	tb.Helper()
	containerClient, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoErrorf(tb, err, "Failed to get provider")

	containers, err := containerClient.ContainerList(context.Background(), container.ListOptions{})
	require.NoErrorf(tb, err, "Failed to list containers")
	require.NotEmptyf(tb, containers, "container list empty")

	containerNameRegexp := regexp.MustCompile(fmt.Sprintf(`^\/?%s(_|-)%s(_|-)\d$`, composeIdentifier, serviceName))
	var containerID string
containerLoop:
	for i := range containers {
		c := containers[i]
		for j := range c.Names {
			if containerNameRegexp.MatchString(c.Names[j]) {
				containerID = c.ID
				break containerLoop
			}
		}
	}

	details, err := containerClient.ContainerInspect(context.Background(), containerID)
	require.NoErrorf(tb, err, "Failed to inspect container")

	for k, v := range present {
		keyVal := k + "=" + v
		assert.Contains(tb, details.Config.Env, keyVal)
	}

	for k, v := range absent {
		keyVal := k + "=" + v
		assert.NotContains(tb, details.Config.Env, keyVal)
	}
}

func checkIfError(t *testing.T, err ExecError) {
	t.Helper()
	require.NoErrorf(t, err.Error, "Failed when running %v", err.Command)

	require.NoErrorf(t, err.Stdout, "An error in Stdout happened when running %v", err.Command)

	require.NoErrorf(t, err.Stderr, "An error in Stderr happened when running %v", err.Command)

	assert.NotNil(t, err.StdoutOutput)
	assert.NotNil(t, err.StderrOutput)
}
