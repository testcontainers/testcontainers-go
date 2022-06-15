package testcontainers

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleNewLocalDockerCompose() {
	path := "/path/to/docker-compose.yml"

	_, _ = NewLocalDockerCompose([]string{path}, "my_project")
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

	compose, err := NewLocalDockerCompose([]string{path}, "my_project")
	if err != nil {
		_ = fmt.Errorf("Failed when creating local docker compose: %v", err)
	}

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

	compose, err := NewLocalDockerCompose([]string{path}, "my_project")
	if err != nil {
		_ = fmt.Errorf("Failed when creating local docker compose: %v", err)
	}

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

	compose, err := NewLocalDockerCompose([]string{path}, "my_project")
	if err != nil {
		_ = fmt.Errorf("Failed when creating local docker compose: %v", err)
	}

	compose.WithCommand([]string{"up", "-d"})
}

func ExampleLocalDockerCompose_WithEnv() {
	path := "/path/to/docker-compose.yml"

	compose, err := NewLocalDockerCompose([]string{path}, "my_project")
	if err != nil {
		_ = fmt.Errorf("Failed when creating local docker compose: %v", err)
	}

	compose.WithEnv(map[string]string{
		"FOO": "foo",
		"BAR": "bar",
	})
}

func TestLocalDockerCompose(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, execErr)
}

func TestDockerComposeStrategyForInvalidService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	assert.NotEqual(t, execErr.Error, nil, "Expected error to be thrown because service with wait strategy is not running")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitLogStrategy(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second).WithOccurrence(1)).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithWaitForService(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WaitForService("nginx_1", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitHTTPStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithContainerName(t *testing.T) {
	path := "./testresources/docker-compose-container-name.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier)
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginxy", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := "./testresources/docker-compose-no-exposed-ports.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService("nginx_1", 9080, wait.ForLog("Configuration complete; ready for start up")).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithMultipleWaitStrategies(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").WithStartupTimeout(10*time.Second)).
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithFailedStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").WithPort("8080/tcp").WithStartupTimeout(5*time.Second)).
		Invoke()
	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.NotEqual(t, execErr.Error, nil, "Expected error to be thrown because of a wrong suplied wait strategy")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestLocalDockerComposeComplex(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestLocalDockerComposeWithEnvironment(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
		}).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")

	containerNameNginx := compose.Identifier + "_nginx_1"

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, containerNameNginx, present, absent)
}

func TestLocalDockerComposeWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		"testresources/docker-compose-simple.yml",
		"testresources/docker-compose-postgres.yml",
		"testresources/docker-compose-override.yml",
	}

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose(composeFiles, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"bar": "BAR",
			"foo": "FOO",
		}).
		Invoke()
	checkIfError(t, execErr)

	assert.Equal(t, 3, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
	assert.Contains(t, compose.Services, "postgres")

	containerNameNginx := compose.Identifier + "_nginx_1"

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, containerNameNginx, present, absent)
}

func TestLocalDockerComposeWithVolume(t *testing.T) {
	path := "./testresources/docker-compose-volume.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	defer func() {
		checkIfError(t, compose.Down())
		assertVolumeDoesNotExist(t, fmt.Sprintf("%s_mydata", identifier))
	}()

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, execErr)
}

func TestLocalDockerComposeWithReaperCleanup(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, execErr)

	time.Sleep(11 * time.Second)
	assertContainerDoesNotExist(t, identifier+"-nginx-1")
}

func TestLocalDockerComposeWithVolumeAndReaperCleanup(t *testing.T) {
	path := "./testresources/docker-compose-volume.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose, err := NewLocalDockerCompose([]string{path}, identifier, WithLogger(TestLogger(t)))
	assert.Nil(t, err)

	execErr := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, execErr)

	time.Sleep(11 * time.Second)
	assertContainerDoesNotExist(t, identifier+"-nginx-1")
	assertVolumeDoesNotExist(t, fmt.Sprintf("%s_mydata", identifier))
}

func assertVolumeDoesNotExist(t *testing.T, volume string) {
	args := []string{"volume", "inspect", volume}

	output, _ := executeAndGetOutput("docker", args)
	if !strings.Contains(output, "No such volume") {
		t.Fatalf("Expected volume %q to not exist", volume)
	}
}

func assertContainerEnvironmentVariables(t *testing.T, containerName string, present map[string]string, absent map[string]string) {
	args := []string{"exec", containerName, "env"}

	output, err := executeAndGetOutput("docker", args)
	checkIfError(t, err)

	for k, v := range present {
		keyVal := k + "=" + v
		assert.Contains(t, output, keyVal)
	}

	for k, v := range absent {
		keyVal := k + "=" + v
		assert.NotContains(t, output, keyVal)
	}
}

func assertContainerDoesNotExist(t *testing.T, containerName string) {
	args := []string{"ps", "-q", "--filter", "name=" + containerName}

	output, err := executeAndGetOutput("docker", args)
	checkIfError(t, err)
	if len(output) > 0 {
		t.Fatalf("Expected container %q to not exist", containerName)
	}
}

func checkIfError(t *testing.T, err ExecError) {
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

func executeAndGetOutput(command string, args []string) (string, ExecError) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()

	return string(out), ExecError{
		Error:        err,
		StderrOutput: out,
		StdoutOutput: out,
	}
}
