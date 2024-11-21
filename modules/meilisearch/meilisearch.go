package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultMasterKey = "just-a-master-key-for-test"
	defaultHTTPPort  = "7700/tcp"
	masterKeyEnvVar  = "MEILI_MASTER_KEY"
)

// MeilisearchContainer represents the Meilisearch container type used in the module
type MeilisearchContainer struct {
	testcontainers.Container
	masterKey string
}

// MasterKey retrieves the master key of the Meilisearch container
func (c *MeilisearchContainer) MasterKey() string {
	return c.masterKey
}

// Run creates an instance of the Meilisearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MeilisearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultHTTPPort},
		Env: map[string]string{
			masterKeyEnvVar: defaultMasterKey,
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(settings)
		}
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	if settings.DumpDataFilePath != "" {
		genericContainerReq.Files = []testcontainers.ContainerFile{
			{
				HostFilePath:      settings.DumpDataFilePath,
				ContainerFilePath: "/dumps/" + settings.DumpDataFileName,
				FileMode:          0o755,
			},
		}
		genericContainerReq.Cmd = []string{"meilisearch", "--import-dump", "/dumps/" + settings.DumpDataFileName}
	}

	// the wait strategy does not support TLS at the moment,
	// so we need to disable it in the strategy for now.
	genericContainerReq.WaitingFor = wait.ForHTTP("/health").
		WithPort(defaultHTTPPort).
		WithTLS(false).
		WithStartupTimeout(120 * time.Second).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			decoder := json.NewDecoder(body)
			r := struct {
				Status string `json:"status"`
			}{}
			if err := decoder.Decode(&r); err != nil {
				return false
			}

			return r.Status == "available"
		})

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MeilisearchContainer
	if container != nil {
		c = &MeilisearchContainer{Container: container, masterKey: req.Env[masterKeyEnvVar]}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// Address retrieves the address of the Meilisearch container.
// It will use http as protocol, as TLS is not supported at the moment.
func (c *MeilisearchContainer) Address(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, defaultHTTPPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	return "http://" + net.JoinHostPort(host, containerPort.Port()), nil
}
