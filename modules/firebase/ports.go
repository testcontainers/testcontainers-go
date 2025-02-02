package firebase

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
)

func (c *Container) ConnectionString(ctx context.Context, portName nat.Port) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, portName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
