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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultHTTPPort),
		testcontainers.WithEnv(map[string]string{
			masterKeyEnvVar: defaultMasterKey,
		}),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(settings); err != nil {
				return nil, fmt.Errorf("meilisearch option: %w", err)
			}
		}
	}

	if settings.DumpDataFilePath != "" {
		moduleOpts = append(moduleOpts, testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      settings.DumpDataFilePath,
			ContainerFilePath: "/dumps/" + settings.DumpDataFileName,
			FileMode:          0o755,
		}))

		moduleOpts = append(moduleOpts, testcontainers.WithCmd("meilisearch", "--import-dump", "/dumps/"+settings.DumpDataFileName))
	}

	// the wait strategy does not support TLS at the moment,
	// so we need to disable it in the strategy for now.
	moduleOpts = append(moduleOpts, testcontainers.WithWaitStrategy(wait.ForHTTP("/health").
		WithPort(defaultHTTPPort).
		WithTLS(false).
		WithStartupTimeout(120*time.Second).
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
		})))

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(settings.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MeilisearchContainer
	if ctr != nil {
		c = &MeilisearchContainer{Container: ctr, masterKey: settings.env[masterKeyEnvVar]}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
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
