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
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
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
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").
			WithTimeout(10*time.Second).
			WithOccurrence(1)).
		Invoke()
	assert.NotEqual(t, err.Error, nil, "Expected error to be thrown because service with wait strategy is not running")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitLogStrategy(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		// Appending with _1 as given in the Java Test-Containers Example
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").
			WithTimeout(10*time.Second).
			WithOccurrence(1)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithWaitHTTPStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

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
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").
			WithPort("80/tcp").
			WithTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithWaitStrategy_NoExposedPorts(t *testing.T) {
	path := "./testresources/docker-compose-no-exposed-ports.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService("nginx_1", 9080, wait.ForLog("Configuration complete; ready for start up")).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestDockerComposeWithMultipleWaitStrategies(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithExposedService("mysql_1", 13306, wait.NewLogStrategy("started").
			WithTimeout(10*time.Second)).
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").
			WithPort("80/tcp").
			WithTimeout(10*time.Second)).
		Invoke()
	checkIfError(t, err)

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")
}

func TestDockerComposeWithFailedStrategy(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

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
		WithExposedService("nginx_1", 9080, wait.NewHTTPStrategy("/").
			WithPort("8080/tcp").
			WithTimeout(5*time.Second)).
		Invoke()
	// Verify that an error is thrown and not nil
	// A specific error message matcher is not asserted since the docker library can change the return message, breaking this test
	assert.NotEqual(t, err.Error, nil, "Expected error to be thrown because of a wrong suplied wait strategy")

	assert.Equal(t, 1, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
}

func TestLocalDockerComposeComplex(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)
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
	path := "./testresources/docker-compose-simple.yml"

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
		Invoke()
	checkIfError(t, err)

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
		"testresources/docker-compose-override.yml",
	}

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose(composeFiles, identifier)
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

	assert.Equal(t, 2, len(compose.Services))
	assert.Contains(t, compose.Services, "nginx")
	assert.Contains(t, compose.Services, "mysql")

	containerNameNginx := compose.Identifier + "_nginx_1"

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, containerNameNginx, present, absent)
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
}

func executeAndGetOutput(command string, args []string) (string, ExecError) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", ExecError{Error: err}
	}

	return string(out), ExecError{Error: nil}
}
