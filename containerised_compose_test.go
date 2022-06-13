package testcontainers

import "testing"

func Test_NewContainerisedDockerCompose(t *testing.T) {
	compose := NewContainerisedDockerCompose([]string{}, "", ContainerisedDockerComposeOptions{})

	res := compose.Invoke()

	if res.Error != nil {
		t.Fatal()
	}
}
