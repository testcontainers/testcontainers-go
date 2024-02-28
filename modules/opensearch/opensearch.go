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

// RunContainer creates an instance of the OpenSearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.11.1",
		ExposedPorts: []string{defaultHTTPPort, "9600/tcp"},
		Env: map[string]string{
			"discovery.type":              "single-node",
			"DISABLE_INSTALL_DEMO_CONFIG": "true",
			"DISABLE_SECURITY_PLUGIN":     "true",
			"OPENSEARCH_USERNAME":         defaultUsername,
			"OPENSEARCH_PASSWORD":         defaultPassword,
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
					Soft: 65536, // Maximum number of open files for the opensearch user - set to at least 65536
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
		opt.Customize(&genericContainerReq)
	}

	// set credentials if they are provided, otherwise use the defaults
	if settings.Username != "" {
		genericContainerReq.Env["OPENSEARCH_USERNAME"] = settings.Username
	}
	if settings.Password != "" {
		genericContainerReq.Env["OPENSEARCH_PASSWORD"] = settings.Password
	}

	username := genericContainerReq.Env["OPENSEARCH_USERNAME"]
	password := genericContainerReq.Env["OPENSEARCH_PASSWORD"]

	// the wat strategy does not support TLS at the moment,
	// so we need to disable it in the strategy for now.
	genericContainerReq.WaitingFor = wait.ForHTTP("/").
		WithPort("9200").
		WithTLS(false).
		WithStartupTimeout(120*time.Second).
		WithStatusCodeMatcher(func(status int) bool {
			return status == 200
		}).
		WithBasicAuth(username, password).
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
		})

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &OpenSearchContainer{Container: container, User: username, Password: password}, nil
}

// Address retrieves the address of the OpenSearch container.
// It will use http as protocol, as TLS is not supported at the moment.
func (c *OpenSearchContainer) Address(ctx context.Context) (string, error) {
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
