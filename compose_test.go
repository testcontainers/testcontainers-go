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

	compose := NewLocalDockerCompose(path, identifier)

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
	path := "./testresources/docker-compose.yml"

	identifier := strings.ToLower(RandomString(6))

	compose := NewLocalDockerCompose(path, identifier)
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

	args := []string{
		"exec", compose.Identifier + "_nginx_1", "env",
	}

	output, err := executeAndGetOutput("docker", args)
	checkIfError(t, err)
	assert.Contains(t, output, "bar=BAR")
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
