package image

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core/mock"
)

// testBuilder is a test helper that implements the Builder interface.
type testBuilder struct {
	Builder
}

// BuildOptions returns a types.ImageBuildOptions for the testBuilder,
// including one single tag in the Tags field.
func (t *testBuilder) BuildOptions() (types.ImageBuildOptions, error) {
	return types.ImageBuildOptions{
		Tags: []string{"test"}, // needed because, the Build function returns the first tag
	}, nil
}

// GetContext returns nil, nil for the testBuilder.
func (t *testBuilder) GetContext() (io.Reader, error) {
	return nil, nil
}

// GetDockerfile returns an empty string for the testBuilder.
func (t *testBuilder) GetDockerfile() string {
	return ""
}

// GetRepo returns an empty string for the testBuilder.
func (t *testBuilder) GetRepo() string {
	return ""
}

// GetTag returns an empty string for the testBuilder.
func (t *testBuilder) GetTag() string {
	return ""
}

// BuildLogWriter returns io.Discard for the testBuilder.
func (t *testBuilder) BuildLogWriter() io.Writer {
	return io.Discard
}

// ShouldBuildImage returns true for the testBuilder.
func (t *testBuilder) ShouldBuildImage() bool {
	return true
}

// GetBuildArgs returns nil for the testBuilder.
func (t *testBuilder) GetBuildArgs() map[string]*string {
	return nil
}

// GetAuthConfigs returns nil for the testBuilder.
func (t *testBuilder) GetAuthConfigs() map[string]registry.AuthConfig {
	return nil
}

func TestDockerProvider_BuildImage_Retries(t *testing.T) {
	tests := []struct {
		name        string
		errReturned error
		shouldRetry bool
	}{
		{
			name:        "success/no-retry",
			errReturned: nil,
			shouldRetry: false,
		},
		{
			name:        "resource-not-found/no-retry",
			errReturned: errdefs.NotFound(errors.New("not available")),
			shouldRetry: false,
		},
		{
			name:        "invalid-parameters/no-retry",
			errReturned: errdefs.InvalidParameter(errors.New("invalid")),
			shouldRetry: false,
		},
		{
			name:        "resource-access-unauthorized/no-retry",
			errReturned: errdefs.Unauthorized(errors.New("not authorized")),
			shouldRetry: false,
		},
		{
			name:        "resource-access-forbidden/no-retry",
			errReturned: errdefs.Forbidden(errors.New("forbidden")),
			shouldRetry: false,
		},
		{
			name:        "not-implemented-by-provider/no-retry",
			errReturned: errdefs.NotImplemented(errors.New("unknown method")),
			shouldRetry: false,
		},
		{
			name:        "system-error/no-retry",
			errReturned: errdefs.System(errors.New("system error")),
			shouldRetry: false,
		},
		{
			name:        "non-permanent-error/retry",
			errReturned: errors.New("whoops"),
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli := mock.NewErrClient(tt.errReturned)

			// give a chance to retry
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, err := buildWithClient(ctx, mockCli, &testBuilder{})
			if tt.errReturned != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Positive(t, mockCli.ImageBuildCount())
			require.Equal(t, tt.shouldRetry, mockCli.ImageBuildCount() > 1)
		})
	}
}
