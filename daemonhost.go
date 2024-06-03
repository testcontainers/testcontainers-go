package testcontainers

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/testcontainers/testcontainers-go/internal/core"
	tcnetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
)

// DaemonHost returns the host of the Docker daemon
func DaemonHost(ctx context.Context) (string, error) {
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
				ip, err = defaultGatewayIP()
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

// see https://github.com/testcontainers/testcontainers-java/blob/main/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L46
func defaultGatewayIP() (string, error) {
	// see https://github.com/testcontainers/testcontainers-java/blob/3ad8d80e2484864e554744a4800a81f6b7982168/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L27
	cmd := exec.Command("sh", "-c", "ip route|awk '/default/ { print $3 }'")
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.New("failed to detect docker host")
	}
	ip := strings.TrimSpace(string(stdout))
	if len(ip) == 0 {
		return "", errors.New("failed to parse default gateway IP")
	}
	return ip, nil
}
