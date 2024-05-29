package testcontainers

import (
	"context"
	"errors"
	"net/url"
	"os"

	"github.com/testcontainers/testcontainers-go/internal/core"
	tcnetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
)

func daemonHost(ctx context.Context) (string, error) {
	var hostCache string

	host, exists := os.LookupEnv("TC_HOST")
	if exists {
		hostCache = host
		return hostCache, nil
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	// infer from Docker host
	url, err := url.Parse(cli.DaemonHost())
	if err != nil {
		return "", err
	}

	switch url.Scheme {
	case "http", "https", "tcp":
		hostCache = url.Hostname()
	case "unix", "npipe":
		if core.InAContainer() {
			ip, err := tcnetwork.GetGatewayIP(ctx)
			if err != nil {
				ip, err = core.DefaultGatewayIP()
				if err != nil {
					ip = "localhost"
				}
			}
			hostCache = ip
		} else {
			hostCache = "localhost"
		}
	default:
		return "", errors.New("could not determine host through env or docker host")
	}

	return hostCache, nil
}
