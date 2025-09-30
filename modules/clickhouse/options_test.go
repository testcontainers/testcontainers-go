package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestWithPassword(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	err := WithPassword("")(&req)
	require.Error(t, err)
}
