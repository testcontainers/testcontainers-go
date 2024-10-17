package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	MasterKey string
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
			return nil, err
		}
	}

	if settings.DumpDataFilePath != "" {
		genericContainerReq.Files = []testcontainers.ContainerFile{
			{
				HostFilePath:      settings.DumpDataFilePath,
				ContainerFilePath: "/dumps/" + settings.DumpDataFileName,
				FileMode:          0o777,
			},
		}
		genericContainerReq.Cmd = []string{"meilisearch", "--import-dump", fmt.Sprintf("/dumps/%s", settings.DumpDataFileName)}
	}

	// the wait strategy does not support TLS at the moment,
	// so we need to disable it in the strategy for now.
	genericContainerReq.WaitingFor = wait.ForHTTP("/health").
		WithPort(defaultHTTPPort).
		WithTLS(false).
		WithStartupTimeout(120 * time.Second).
		WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			bs, err := io.ReadAll(body)
			if err != nil {
				return false
			}

			type response struct {
				Status string `json:"status"`
			}

			var r response
			err = json.Unmarshal(bs, &r)
			if err != nil {
				return false
			}

			return r.Status == "available"
		})

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MeilisearchContainer
	if container != nil {
		c = &MeilisearchContainer{Container: container, MasterKey: req.Env[masterKeyEnvVar]}
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
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + net.JoinHostPort(host, containerPort.Port()), nil
}

// WithMasterKey sets the master key for the Meilisearch container
// it satisfies the testcontainers.ContainerCustomizer interface
func WithMasterKey(masterKey string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MEILI_MASTER_KEY"] = masterKey
		return nil
	}
}
