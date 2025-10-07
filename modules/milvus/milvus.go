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
	etcdPort               = "2379/tcp"
	httpPort               = "9091/tcp"
	grpcPort               = "19530/tcp"
)

// MilvusContainer represents the Milvus container type used in the module
type MilvusContainer struct {
	testcontainers.Container
}

// ConnectionString returns the connection string for the milvus container, using the default 19530 port, and
// obtaining the host and exposed port from the container.
func (c *MilvusContainer) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, grpcPort, "")
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

	// Adapted from https://github.com/milvus-io/milvus/blob/v2.6.3/scripts/standalone_embed.sh
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(grpcPort, httpPort, etcdPort),
		testcontainers.WithEnv(map[string]string{
			"ETCD_USE_EMBED":     "true",
			"ETCD_DATA_DIR":      "/var/lib/milvus/etcd",
			"ETCD_CONFIG_PATH":   embedEtcdContainerPath,
			"COMMON_STORAGETYPE": "local",
			"DEPLOY_MODE":        "STANDALONE",
		}),
		testcontainers.WithCmd("milvus", "run", "standalone"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForHTTP("/healthz").
				WithPort(httpPort).
				WithStartupTimeout(time.Minute).
				WithPollInterval(time.Second),
			wait.ForListeningPort(httpPort).
				WithStartupTimeout(time.Minute),
			wait.ForListeningPort(grpcPort).
				WithStartupTimeout(time.Minute),
		)),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			ContainerFilePath: embedEtcdContainerPath,
			Reader:            config,
		}),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MilvusContainer
	if ctr != nil {
		c = &MilvusContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run milvus: %w", err)
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
