package testcontainers

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalDockerCompose(t *testing.T) {
	path := "./testresources/docker-compose.yml"

	identifier := strings.ToLower(RandomString(6))

	compose := NewLocalDockerCompose([]string{path}, identifier)

	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	checkIfError(t, err)

	assertContainerEnvContainsKeyValue(t, compose.Identifier+"_nginx_1", "bar", "")

	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()
}

func TestLocalDockerComposeWithEnvironment(t *testing.T) {
	path := "./testresources/docker-compose.yml"

	identifier := strings.ToLower(RandomString(6))

	compose := NewLocalDockerCompose([]string{path}, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"foo": "BAR",
		}).
		Invoke()
	checkIfError(t, err)

	assertContainerEnvContainsKeyValue(t, compose.Identifier+"_nginx_1", "bar", "BAR")
}

func TestLocalDockerComposeWithMultipleComposeFiles(t *testing.T) {
	composeFiles := []string{
		"testresources/docker-compose.yml",
		"testresources/docker-compose-override.yml",
	}

	identifier := strings.ToLower(RandomString(6))

	compose := NewLocalDockerCompose(composeFiles, identifier)
	destroyFn := func() {
		err := compose.Down()
		checkIfError(t, err)
	}
	defer destroyFn()

	err := compose.
		WithCommand([]string{"up", "-d"}).
		WithEnv(map[string]string{
			"foo": "BAR",
			"bar": "FOO",
		}).
		Invoke()
	checkIfError(t, err)

	assertContainerEnvContainsKeyValue(t, compose.Identifier+"_nginx_1", "foo", "FOO")
}

func assertContainerEnvContainsKeyValue(t *testing.T, identifier string, key string, value string) {
	args := []string{"exec", identifier, "env"}

	output, err := executeAndGetOutput("docker", args)
	checkIfError(t, err)

	keyVal := key + "=" + value
	assert.Contains(t, output, keyVal)
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
