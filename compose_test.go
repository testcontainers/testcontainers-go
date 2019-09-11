package testcontainers

import "testing"

func TestLocalDockerCompose(t *testing.T) {
	path := "./testresources/docker-compose.yml"

	identifier := RandomString(6)

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

func checkIfError(t *testing.T, err ExecError) {
	if err.Error != nil || err.Stdout != nil || err.Stderr != nil {
		t.Fatal(err)
	}
}
