package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/exec"
)

func TestWithStartupCommand_unit(t *testing.T) {
	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{},
	}

	testExec := NewRawCommand([]string{"touch", ".testcontainers"}, exec.WithWorkingDir("/tmp"))

	err := WithStartupCommand(testExec)(&req)
	require.NoError(t, err)

	require.Len(t, req.LifecycleHooks, 1)
	require.Len(t, req.LifecycleHooks[0].PostStarts, 1)
}

func TestWithAfterReadyCommand_unit(t *testing.T) {
	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{},
	}

	testExec := NewRawCommand([]string{"touch", "/tmp/.testcontainers"})

	err := WithAfterReadyCommand(testExec)(&req)
	require.NoError(t, err)

	require.Len(t, req.LifecycleHooks, 1)
	require.Len(t, req.LifecycleHooks[0].PostReadies, 1)
}
