package k3s

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v3"
)

var (
	// containerPorts {
	defaultKubeSecurePort     = "6443/tcp"
	defaultRancherWebhookPort = "8443/tcp"
	// }
	defaultKubeConfigK3sPath = "/etc/rancher/k3s/k3s.yaml"
)

// K3sContainer represents the K3s container type used in the module
type K3sContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the K3s container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K3sContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "docker.io/rancher/k3s:v1.27.1-k3s1",
		ExposedPorts: []string{
			defaultKubeSecurePort,
			defaultRancherWebhookPort,
		},
		Privileged: true,
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CgroupnsMode = "host"
			hc.Tmpfs = map[string]string{
				"/run":     "",
				"/var/run": "",
			}
			hc.Mounts = []mount.Mount{}

		},
		Cmd: []string{
			"server",
			"--disable=traefik",
			"--tls-san=localhost",
		},
		Env: map[string]string{
			"K3S_KUBECONFIG_MODE": "644",
		},
		WaitingFor: wait.ForLog("k3s is up and running"),
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

	return &K3sContainer{Container: container}, nil
}

// GetKubeConfig returns the modified kubeconfig with server url
func (c *K3sContainer) GetKubeConfig(ctx context.Context) ([]byte, error) {
	hostIP, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hostIP: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, nat.Port(defaultKubeSecurePort))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	reader, err := c.CopyFileFromContainer(ctx, defaultKubeConfigK3sPath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file from container: %w", err)
	}

	kubeConfigYaml, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from container: %w", err)
	}

	server := "https://" + fmt.Sprintf("%v:%d", hostIP, mappedPort.Int())
	newKubeConfig, err := kubeConfigWithServerUrl(string(kubeConfigYaml), server)
	if err != nil {
		return nil, fmt.Errorf("failed to modify kubeconfig with server url: %w", err)
	}

	return newKubeConfig, nil
}

func kubeConfigWithServerUrl(kubeConfigYaml, server string) ([]byte, error) {

	kubeConfig, err := unmarshal([]byte(kubeConfigYaml))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal kubeconfig: %w", err)
	}

	kubeConfig.Clusters[0].Cluster.Server = server
	modifiedKubeConfig, err := marshal(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	return modifiedKubeConfig, nil
}

func marshal(config *KubeConfigValue) ([]byte, error) {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func unmarshal(bytes []byte) (*KubeConfigValue, error) {
	var kubeConfig KubeConfigValue
	err := yaml.Unmarshal(bytes, &kubeConfig)
	if err != nil {
		return nil, err
	}
	return &kubeConfig, nil
}
