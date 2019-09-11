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
		compose.WithCommand([]string{"down"})

		err := compose.Invoke()
		checkIfError(t, err)
	}
	defer destroyFn()
}

func checkIfError(t *testing.T, err execError) {
	if err.Error != nil || err.Stdout != nil || err.Stderr != nil {
		t.Fatal(err)
	}
}
