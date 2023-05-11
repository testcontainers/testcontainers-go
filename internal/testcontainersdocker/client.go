package testcontainersdocker

import (
	"context"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/testcontainerssession"
)

// NewClient returns a new docker client with the default options
// reading the Testcontainers configuration from the Testcontainers file
func NewClient(ctx context.Context, ops ...client.Opt) (*client.Client, error) {
	tcConfig := config.Read(ctx)

	host := tcConfig.Host

	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	if host != "" {
		opts = append(opts, client.WithHost(host))

		// for further information, read https://docs.docker.com/engine/security/protect-access/
		if tcConfig.TLSVerify == 1 {
			cacertPath := filepath.Join(tcConfig.CertPath, "ca.pem")
			certPath := filepath.Join(tcConfig.CertPath, "cert.pem")
			keyPath := filepath.Join(tcConfig.CertPath, "key.pem")

			opts = append(opts, client.WithTLSClientConfig(cacertPath, certPath, keyPath))
		}
	}

	opts = append(opts, client.WithHTTPHeaders(
		map[string]string{
			"x-tc-sid": testcontainerssession.String(),
		}),
	)

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	_, err = cli.Ping(context.TODO())
	if err != nil {
		// fallback to environment
		cli, err = defaultClient(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return cli, nil
}

// defaultClient returns a plain, new docker client with the default options
func defaultClient(ctx context.Context, ops ...client.Opt) (*client.Client, error) {
	if len(ops) == 0 {
		ops = []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	}

	return client.NewClientWithOpts(ops...)
}
