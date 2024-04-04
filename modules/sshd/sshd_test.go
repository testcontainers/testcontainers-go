package sshd_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainersshd "github.com/testcontainers/testcontainers-go/modules/sshd"
)

func TestSSHD(t *testing.T) {
	ctx := context.Background()
	opts := []testcontainers.ContainerCustomizer{}
	_, err := testcontainersshd.RunContainer(ctx, opts...)
	require.NoError(t, err)

	// @TODO: Test the forwarding container <> host
}
