package testcontainers

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const containerNameNginx = "nginx-simple"

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
