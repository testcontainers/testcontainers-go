package network

import (
	"context"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

type Network interface {
	Remove(context.Context) error
}

// DockerNetwork represents a network started using Docker
type DockerNetwork struct {
	ID                string // Network ID from Docker
	Driver            string
	Name              string
	terminationSignal chan bool
}

// Remove is used to remove the network by its ID. It is usually triggered by as defer function.
func (n *DockerNetwork) Remove(ctx context.Context) error {
	select {
	// close reaper if it was created
	case n.terminationSignal <- true:
	default:
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.NetworkRemove(ctx, n.ID)
}

func (n *DockerNetwork) SetTerminationSignal(signal chan bool) {
	n.terminationSignal = signal
}
