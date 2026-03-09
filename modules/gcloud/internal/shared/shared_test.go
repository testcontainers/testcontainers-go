package shared_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/shared"
)

func TestDefaultOptions(t *testing.T) {
	opts := shared.DefaultOptions()
	require.Equal(t, shared.DefaultProjectID, opts.ProjectID)
}

func TestWithProjectID(t *testing.T) {
	opts := shared.DefaultOptions()

	err := shared.WithProjectID("test-project")(&opts)
	require.NoError(t, err)
	require.Equal(t, "test-project", opts.ProjectID)
}

func TestCustomize(t *testing.T) {
	err := shared.WithProjectID("test-project-2").Customize(&testcontainers.GenericContainerRequest{})
	require.NoError(t, err)
}
