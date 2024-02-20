package milvus

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mounts/embedEtcd.yaml.tpl
var embedEtcdConfigTpl string

const embedEtcdContainerPath string = "/milvus/configs/embedEtcd.yaml"

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
	port, err := c.MappedPort(ctx, "19530/tcp")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// RunContainer creates an instance of the Milvus container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MilvusContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "milvusdb/milvus:v2.3.9",
		ExposedPorts: []string{"19530/tcp", "9091/tcp", "2379/tcp"},
		Env: map[string]string{
			"ETCD_USE_EMBED":     "true",
			"ETCD_DATA_DIR":      "/var/lib/milvus/etcd",
			"ETCD_CONFIG_PATH":   embedEtcdContainerPath,
			"COMMON_STORAGETYPE": "local",
		},
		Cmd:        []string{"milvus", "run", "standalone"},
		WaitingFor: wait.ForHTTP("/healthz").WithPort("9091").WithStartupTimeout(60 * time.Second).WithPollInterval(30 * time.Second),
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostCreates: []testcontainers.ContainerHook{
					// Copy the default embed etcd config to container after it's created.
					// Otherwise the milvus container will panic on startup.
					createDefaultEmbedEtcdConfig,
				},
				PostStarts: []testcontainers.ContainerHook{regenerateEmbedEtcdConfig},
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &MilvusContainer{Container: container}, nil
}

type embedEtcdConfigTplParams struct {
	Port int
}

func renderEmbedEtcdConfig(port int) ([]byte, error) {
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

	return embedEtcdYaml.Bytes(), nil
}

// createDefaultEmbedEtcdConfig creates a default embed etcd config file,
// using the default port 2379 as the advertised port. The file is then copied to the container.
func createDefaultEmbedEtcdConfig(ctx context.Context, c testcontainers.Container) error {
	// Otherwise the milvus container will panic on startup.
	defaultEmbedEtcdConfig, err := renderEmbedEtcdConfig(2379)
	if err != nil {
		return fmt.Errorf("failed to render default config: %w", err)
	}

	tmpDir := os.TempDir()
	defaultEmbedEtcdConfigPath := fmt.Sprintf("%s/embedEtcd.yaml", tmpDir)

	if err := os.WriteFile(defaultEmbedEtcdConfigPath, defaultEmbedEtcdConfig, 0o644); err != nil {
		return fmt.Errorf("failed to write default embed etcd config to a temporary dir: %w", err)
	}

	if err != nil {
		return fmt.Errorf("can't create default embed etcd config: %w", err)
	}
	defer os.Remove(defaultEmbedEtcdConfigPath)

	err = c.CopyFileToContainer(ctx, defaultEmbedEtcdConfigPath, embedEtcdContainerPath, 0o644)
	if err != nil {
		return fmt.Errorf("can't copy %s to container: %w", defaultEmbedEtcdConfigPath, err)
	}

	return nil
}

// regenerateEmbedEtcdConfig updates the embed etcd config file with the mapped port
func regenerateEmbedEtcdConfig(ctx context.Context, c testcontainers.Container) error {
	containerPort, err := c.MappedPort(ctx, "2379/tcp")
	if err != nil {
		return fmt.Errorf("failed to get mapped port: %w", err)
	}

	embedEtcdConfig, err := renderEmbedEtcdConfig(containerPort.Int())
	if err != nil {
		return fmt.Errorf("failed to embed etcd config: %w", err)
	}

	err = c.CopyToContainer(ctx, embedEtcdConfig, embedEtcdContainerPath, 600)
	if err != nil {
		return fmt.Errorf("failed to copy embedEtcd.yaml into container: %w", err)
	}

	return nil
}
