package opensearch

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
	defaultPassword = "admin"
	defaultUsername = "admin"
	defaultHTTPPort = "9200/tcp"
)

// OpenSearchContainer represents the OpenSearch container type used in the module
type OpenSearchContainer struct {
	testcontainers.Container
	User     string
	Password string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the OpenSearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error) {
	return Run(ctx, "opensearchproject/opensearch:2.11.1", opts...)
}

// Run creates an instance of the OpenSearch container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultHTTPPort, "9600/tcp"),
		testcontainers.WithEnv(map[string]string{
			"discovery.type":              "single-node",
			"DISABLE_INSTALL_DEMO_CONFIG": "true",
			"DISABLE_SECURITY_PLUGIN":     "true",
			"OPENSEARCH_USERNAME":         defaultUsername,
			"OPENSEARCH_PASSWORD":         defaultPassword,
		}),
		testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			hc.Ulimits = []*units.Ulimit{
				{
					Name: "memlock",
					Soft: -1, // Set memlock to unlimited (no soft or hard limit)
					Hard: -1,
				},
				{
					Name: "nofile",
					Soft: 65536, // Maximum number of open files for the opensearch user - set to at least 65536
					Hard: 65536,
				},
			}
		}),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(settings)
		}
	}

	// set credentials if they are provided, otherwise use the defaults
	moduleOpts = append(moduleOpts, testcontainers.WithEnv(map[string]string{
		"OPENSEARCH_USERNAME": settings.Username,
		"OPENSEARCH_PASSWORD": settings.Password,
	}))

	// the wat strategy does not support TLS at the moment,
	// so we need to disable it in the strategy for now.
	moduleOpts = append(moduleOpts, testcontainers.WithAdditionalWaitStrategy(
		wait.ForHTTP("/").
			WithPort("9200").
			WithTLS(false).
			WithStartupTimeout(120*time.Second).
			WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}).
			WithBasicAuth(settings.Username, settings.Password).
			WithResponseMatcher(func(body io.Reader) bool {
				bs, err := io.ReadAll(body)
				if err != nil {
					return false
				}

				type response struct {
					Tagline string `json:"tagline"`
				}

				var r response
				err = json.Unmarshal(bs, &r)
				if err != nil {
					return false
				}

				return r.Tagline == "The OpenSearch Project: https://opensearch.org/"
			}),
	),
	)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *OpenSearchContainer
	if ctr != nil {
		c = &OpenSearchContainer{Container: ctr, User: settings.Username, Password: settings.Password}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// Address retrieves the address of the OpenSearch container.
// It will use http as protocol, as TLS is not supported at the moment.
func (c *OpenSearchContainer) Address(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, defaultHTTPPort, "http")
}
