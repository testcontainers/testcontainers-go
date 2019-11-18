package testcontainers

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLocalDockerCompose(t *testing.T) {
	path := "./testresources/docker-compose-simple.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)

	present := map[string]string{
		"bar": "",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier+"_nginx_1", present, absent)

	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()
}

func TestLocalDockerComposeComplex(t *testing.T) {
	path := "./testresources/docker-compose-complex.yml"

	identifier := strings.ToLower(uuid.New().String())

	compose := NewLocalDockerCompose([]string{path}, identifier)

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)

	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()
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

	present := map[string]string{
		"bar": "BAR",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier+"_nginx_1", present, absent)
}

func TestLocalDockerComposeWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		"testresources/docker-compose.yml",
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

	present := map[string]string{
		"bar": "BAR",
		"foo": "FOO",
	}
	absent := map[string]string{}
	assertContainerEnvironmentVariables(t, compose.Identifier+"_nginx_1", present, absent)
}

func assertContainerEnvironmentVariables(t *testing.T, identifier string, present map[string]string, absent map[string]string) {
	args := []string{"exec", identifier, "env"}

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
	if err.Error != nil || err.Stdout != nil || err.Stderr != nil {
		t.Fatal(err)
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
