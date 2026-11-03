package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/moby/moby/client"

	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core/bootstrap"
)

// NewClient returns a new docker client extracting the docker host from the different alternatives
func NewClient(ctx context.Context, ops ...client.Opt) (*client.Client, error) {
	dockerHost, err := ExtractDockerHost(ctx)
	if err != nil {
		return nil, err
	}

	tcConfig := config.Read()
	runtime, err := Runtime(ctx)
	if err != nil {
		return nil, err
	}

	opts := []client.Opt{client.FromEnv}
	if runtime.DockerHost != "" {
		// A managed runtime owns its endpoint; host Docker TLS environment settings
		// must not change the connection to that runtime.
		opts = nil
	}
	if dockerHost != "" {
		opts = append(opts, client.WithHost(dockerHost))

		// for further information, read https://docs.docker.com/engine/security/protect-access/
		if tcConfig.TLSVerify == 1 && runtime.DockerHost == "" {
			cacertPath := filepath.Join(tcConfig.CertPath, "ca.pem")
			certPath := filepath.Join(tcConfig.CertPath, "cert.pem")
			keyPath := filepath.Join(tcConfig.CertPath, "key.pem")

			opts = append(opts, client.WithTLSClientConfig(cacertPath, certPath, keyPath))
		}
	}

	opts = append(opts, client.WithHTTPHeaders(
		map[string]string{
			"x-tc-pp":    bootstrap.ProjectPath(),
			"x-tc-sid":   tcConfig.SessionID,
			"User-Agent": "tc-go/" + internal.Version,
		}),
	)

	// passed options have priority over the default ones
	opts = append(opts, ops...)

	cli, err := client.New(opts...)
	if err != nil {
		return nil, err
	}
	if runtime.DockerHost != "" && cli.DaemonHost() != runtime.DockerHost {
		_ = cli.Close()
		return nil, fmt.Errorf("client endpoint conflicts with provider %q", runtime.Name)
	}

	return cli, nil
}
