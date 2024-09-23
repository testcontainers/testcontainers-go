package milvus

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mounts/embedEtcd.yaml.tpl
var embedEtcdConfigTpl string

const (
	embedEtcdContainerPath = "/milvus/configs/embedEtcd.yaml"
	defaultClientPort      = 2379
	grpcPort               = "19530/tcp"
)

// MilvusContainer represents the Milvus container type used in the module
type MilvusContainer struct {
	testcontainers.Container
}

// ConnectionString returns the connection string for the milvus container, using the default 19530 port, and
// obtaining the host and exposed port from the container.
func (c *MilvusContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, grpcPort)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Milvus container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MilvusContainer, error) {
	return Run(ctx, "milvusdb/milvus:v2.3.9", opts...)
}

// Run creates an instance of the Milvus container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MilvusContainer, error) {
	config, err := renderEmbedEtcdConfig(defaultClientPort)
	if err != nil {
		return nil, fmt.Errorf("render config: %w", err)
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"19530/tcp", "9091/tcp", "2379/tcp"},
		Env: map[string]string{
			"ETCD_USE_EMBED":     "true",
			"ETCD_DATA_DIR":      "/var/lib/milvus/etcd",
			"ETCD_CONFIG_PATH":   embedEtcdContainerPath,
			"COMMON_STORAGETYPE": "local",
		},
		Cmd: []string{"milvus", "run", "standalone"},
		WaitingFor: wait.ForHTTP("/healthz").
			WithPort("9091").
			WithStartupTimeout(time.Minute).
			WithPollInterval(time.Second),
		Files: []testcontainers.ContainerFile{
			{ContainerFilePath: embedEtcdContainerPath, Reader: config},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MilvusContainer
	if container != nil {
		c = &MilvusContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

type embedEtcdConfigTplParams struct {
	Port int
}

// renderEmbedEtcdConfig renders the embed etcd config template with the given port
// and returns it as an io.Reader.
func renderEmbedEtcdConfig(port int) (io.Reader, error) {
	tplParams := embedEtcdConfigTplParams{
		Port: port,
	}

	etcdCfgTpl, err := template.New("embedEtcd.yaml").Parse(embedEtcdConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embed etcd config file template: %w", err)
	}

	var embedEtcdYaml bytes.Buffer
	if err := etcdCfgTpl.Execute(&embedEtcdYaml, tplParams); err != nil {
		return nil, fmt.Errorf("failed to render embed etcd config template: %w", err)
	}

	return &embedEtcdYaml, nil
}
