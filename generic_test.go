package testcontainers

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func TestGenericReusableContainer(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericReusableContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer n1.Terminate(ctx)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)

	if err != nil {
		t.Fatal(err)
	}

	n2, err := GenericReusableContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	c, _, err := n2.Exec(ctx, []string{"bash", copiedFileName})
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatalf("File %s should exist, expected return code 0, got %v", copiedFileName, c)
	}
}
