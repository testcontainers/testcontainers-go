package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultMasterKey = "just-a-master-key-for-test"
	defaultHTTPPort  = "7700/tcp"
)

// MeilisearchContainer represents the Meilisearch container type used in the module
type MeilisearchContainer struct {
	testcontainers.Container
	MasterKey string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Meilisearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MeilisearchContainer, error) {
	return Run(ctx, "getmeili/meilisearch:v1.10.3", opts...)
}

// Run creates an instance of the Meilisearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MeilisearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultHTTPPort},
		Env: map[string]string{
			"MEILI_MASTER_KEY": defaultMasterKey,
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Ulimits = []*units.Ulimit{
				{
					Name: "memlock",
					Soft: -1, // Set memlock to unlimited (no soft or hard limit)
					Hard: -1,
				},
				{
					Name: "nofile",
					Soft: 65536, // Maximum number of open files for the meilisearch user - set to at least 65536
					Hard: 65536,
				},
			}
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

	if settings.DumpDataFileDir != "" {

		genericContainerReq.HostConfigModifier = func(hc *container.HostConfig) {
			//hc.Mounts = append(hc.Mounts, mount.Mount{
			//	Type:     mount.TypeBind,
			//	Source:   settings.DumpDataFileDir,
			//	Target:   "/dumps",
			//	ReadOnly: false,
			//})
			hc.Binds = []string{fmt.Sprintf("/Users/mashail/Projects/testcontainers-go/modules/meilisearch/testdata:/dumps")}
		}
		//genericContainerReq.Mounts = testcontainers.ContainerMounts{
		//	{
		//		Source: testcontainers.GenericVolumeMountSource{
		//			Name: settings.DumpDataFileDir,
		//		},
		//		Target: "/dumps",
		//	},
		//}

		genericContainerReq.Cmd = append(genericContainerReq.Cmd, "meilisearch", "--import-dump", fmt.Sprintf("/dumps/%s", settings.DumpDataFileName))
		//genericContainerReq.Entrypoint = []string{"/bin/sh", "-c", "meilisearch"}
	}

	// the wat strategy does not support TLS at the moment,
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
		c = &MeilisearchContainer{Container: container, MasterKey: req.Env["MEILI_MASTER_KEY"]}
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

	return fmt.Sprintf("http://%s:%s", host, containerPort.Port()), nil
}

// WithMasterKey sets the master key for the Meilisearch container
func WithMasterKey(masterKey string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MEILI_MASTER_KEY"] = masterKey
		return nil
	}
}
