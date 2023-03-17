package testcontainersdocker

import (
	"context"

	"github.com/docker/docker/client"
)

// NewClient returns a new docker client with the default options
func NewClient(ctx context.Context, ops ...client.Opt) (*client.Client, error) {
	if len(ops) == 0 {
		ops = []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	}

	return client.NewClientWithOpts(ops...)
}
