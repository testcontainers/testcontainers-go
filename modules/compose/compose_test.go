package compose

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var complexComposeTestFile string = filepath.Join("testdata", "docker-compose-complex.yml")
var simpleComposeTestFile string = filepath.Join("testdata", "docker-compose-simple.yml")

func ExampleNewLocalDockerCompose() {
	path := "/path/to/docker-compose.yml"

	_ = NewLocalDockerCompose([]string{path}, "my_project")
}

func ExampleLocalDockerCompose() {
	_ = LocalDockerCompose{
		Executable: "docker-compose",
		ComposeFilePaths: []string{
			"/path/to/docker-compose.yml",
			"/path/to/docker-compose-1.yml",
			"/path/to/docker-compose-2.yml",
			"/path/to/docker-compose-3.yml",
		},
		Identifier: "my_project",
		Cmd: []string{
			"up", "-d",
		},
		Env: map[string]string{
			"FOO": "foo",
			"BAR": "bar",
		},
	}
}

func ExampleLocalDockerCompose_Down() {
	path := "/path/to/docker-compose.yml"

	compose := NewLocalDockerCompose([]string{path}, "my_project")

	execError := compose.WithCommand([]string{"up", "-d"}).Invoke()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}

	execError = compose.Down()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}
}

func ExampleLocalDockerCompose_Invoke() {
	path := "/path/to/docker-compose.yml"

	compose := NewLocalDockerCompose([]string{path}, "my_project")

	execError := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Invoke()
	if execError.Error != nil {
		_ = fmt.Errorf("Failed when running: %v", execError.Command)
	}
}

func ExampleLocalDockerCompose_WithCommand() {
	path := "/path/to/docker-compose.yml"

	compose := NewLocalDockerCompose([]string{path}, "my_project")

	compose.WithCommand([]string{"up", "-d"})
}

func ExampleLocalDockerCompose_WithEnv() {
	path := "/path/to/docker-compose.yml"

	compose := NewLocalDockerCompose([]string{path}, "my_project")

	compose.WithEnv(map[string]string{
		"FOO": "foo",
		"BAR": "bar",
	})
}

func TestLocalDockerCompose(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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
func TestDockerComposeStrategyForInvalidService(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService(compose.Format("mysql", "1"), 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	assert.NotEqual(t, err.Error, nil, "Expected error to be thrown because service with wait strategy is not running")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitLogStrategy(t *testing.T) {
	path := complexComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService(compose.Format("mysql", "1"), 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithWaitForService(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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
		WaitForService(compose.Format("nginx", "1"), wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitForShortLifespanService(t *testing.T) {
	path := filepath.Join("testdata", "docker-compose-short-lifespan.yml")

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		//Assumption: tzatziki service wait logic will run before falafel, so that falafel service will exit before
		WaitForService(compose.Format("tzatziki", "1"), wait.ForExit().WithExitTimeout(10*time.Second)).
		WaitForService(compose.Format("falafel", "1"), wait.ForExit().WithExitTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "falafel")
	assert.Contains(t, compose.Services, "tzatziki")
}

func TestDockerComposeWithWaitHTTPStrategy(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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
		WithExposedService(compose.Format("nginx", "1"), 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithContainerName(t *testing.T) {
	path := filepath.Join("testdata", "docker-compose-container-name.yml")

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
		WithExposedService("nginxy", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := filepath.Join("testdata", "docker-compose-no-exposed-ports.yml")

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService(compose.Format("nginx", "1"), 9080, wait.ForLog("Configuration complete; ready for start up")).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithMultipleWaitStrategies(t *testing.T) {
	path := complexComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService(compose.Format("mysql", "1"), 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WithExposedService(compose.Format("nginx", "1"), 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithFailedStrategy(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Invoke()
	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.NotEqual(t, err.Error, nil, "Expected error to be thrown because of a wrong suplied wait strategy")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestLocalDockerComposeComplex(t *testing.T) {
	path := complexComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestLocalDockerComposeWithEnvironment(t *testing.T) {
	path := simpleComposeTestFile

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier, "nginx", present, absent)
}

func TestLocalDockerComposeWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		simpleComposeTestFile,
		filepath.Join("testdata", "docker-compose-postgres.yml"),
		filepath.Join("testdata", "docker-compose-override.yml"),
	}

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose(composeFiles, identifier, WithLogger(testcontainers.TestLogger(t)))
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

	assert.Equal(t, 3, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
	assert.Contains(t, compose.Services, "postgres")

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier, "nginx", present, absent)
}

func TestLocalDockerComposeWithVolume(t *testing.T) {
	path := filepath.Join("testdata", "docker-compose-volume.yml")

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier, WithLogger(testcontainers.TestLogger(t)))
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
	containerClient, err := testcontainers.NewDockerClient()
	if err != nil {
		tb.Fatalf("Failed to get provider: %v", err)
	}

	volumeList, err := containerClient.VolumeList(context.Background(), filters.NewArgs(filters.Arg("name", volumeName)))
	if err != nil {
		tb.Fatalf("Failed to list volumes: %v", err)
	}

	if len(volumeList.Warnings) > 0 {
		tb.Logf("Volume list warnings: %v", volumeList.Warnings)
	}

	if len(volumeList.Volumes) > 0 {
		tb.Fatalf("Volume list is not empty")
	}
}

func assertContainerEnvironmentVariables(
	tb testing.TB,
	composeIdentifier, serviceName string,
	present map[string]string,
	absent map[string]string,
) {
	containerClient, err := testcontainers.NewDockerClient()
	if err != nil {
		tb.Fatalf("Failed to get provider: %v", err)
	}

	containers, err := containerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		tb.Fatalf("Failed to list containers: %v", err)
	} else if len(containers) == 0 {
		tb.Fatalf("container list empty")
	}

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
	if err != nil {
		tb.Fatalf("Failed to inspect container: %v", err)
	}

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
	if err.Error != nil {
		t.Fatalf("Failed when running %v: %v", err.Command, err.Error)
	}

	if err.Stdout != nil {
		t.Fatalf("An error in Stdout happened when running %v: %v", err.Command, err.Stdout)
	}

	if err.Stderr != nil {
		t.Fatalf("An error in Stderr happened when running %v: %v", err.Command, err.Stderr)
	}

	assert.NotNil(t, err.StdoutOutput)
	assert.NotNil(t, err.StderrOutput)
}
