package mock

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// ErrClient is a mock implementation of client.APIClient, which is handy for simulating
// error returns in retry scenarios.
type ErrClient struct {
	client.APIClient

	err                error
	imageBuildCount    int
	containerListCount int
	imagePullCount     int
}

// NewErrClient returns a new ErrClient with the given error.
func NewErrClient(err error) *ErrClient {
	return &ErrClient{err: err}
}

// ImageBuildCount returns the number of times ImageBuild has been called.
func (f *ErrClient) ImageBuildCount() int {
	return f.imageBuildCount
}

// ContainerListCount returns the number of times ContainerList has been called.
func (f *ErrClient) ContainerListCount() int {
	return f.containerListCount
}

// ImagePullCount returns the number of times ImagePull has been called.
func (f *ErrClient) ImagePullCount() int {
	return f.imagePullCount
}

// ImageBuild returns a mock implementation of client.APIClient.ImageBuild.
func (f *ErrClient) ImageBuild(_ context.Context, _ io.Reader, _ types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	f.imageBuildCount++
	return types.ImageBuildResponse{Body: io.NopCloser(&bytes.Buffer{})}, f.err
}

// ContainerList returns a mock implementation of client.APIClient.ContainerList.
func (f *ErrClient) ContainerList(_ context.Context, _ container.ListOptions) ([]types.Container, error) {
	f.containerListCount++
	return []types.Container{{}}, f.err
}

// ImagePull returns a mock implementation of client.APIClient.ImagePull.
func (f *ErrClient) ImagePull(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
	f.imagePullCount++
	return io.NopCloser(&bytes.Buffer{}), f.err
}

// Close returns a mock implementation of client.APIClient.Close.
func (f *ErrClient) Close() error {
	return nil
}
