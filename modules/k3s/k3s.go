package k3s

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"gopkg.in/yaml.v3"

	"github.com/testcontainers/testcontainers-go"
	tcimage "github.com/testcontainers/testcontainers-go/image"
	tclog "github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// containerPorts {
	defaultKubeSecurePort     = "6443/tcp"
	defaultRancherWebhookPort = "8443/tcp"
	// }
	defaultKubeConfigK3sPath = "/etc/rancher/k3s/k3s.yaml"
)

// Container represents the K3s container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}

// path to the k3s manifests directory
const k3sManifests = "/var/lib/rancher/k3s/server/manifests/"

// WithManifest loads the manifest into the cluster. K3s applies it automatically during the startup process
func WithManifest(manifestPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
		manifest := filepath.Base(manifestPath)
		target := k3sManifests + manifest

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      manifestPath,
			ContainerFilePath: target,
		})

		return nil
	}
}

// Run creates an instance of the K3s container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	host, err := getContainerHost(ctx, opts...)
	if err != nil {
		return nil, err
	}

	req := testcontainers.Request{
		Image: img,
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
			"--tls-san=" + host, // Host which will be used to access the Kubernetes server from tests.
		},
		Env: map[string]string{
			"K3S_KUBECONFIG_MODE": "644",
		},
		WaitingFor: wait.ForLog(".*Node controller sync successful.*").AsRegexp(),
		Started:    true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	ctr, err := testcontainers.Run(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{DockerContainer: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func getContainerHost(ctx context.Context, opts ...testcontainers.RequestCustomizer) (string, error) {
	// Use a dummy request to get the provider from options.
	var req testcontainers.Request
	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return "", err
		}
	}

	if req.Logger == nil {
		req.Logger = tclog.StandardLogger()
	}

	daemonHost, err := testcontainers.DaemonHost(ctx)
	if err != nil {
		// Fall back to localhost.
		return "localhost", nil
	}

	return daemonHost, nil
}

// GetKubeConfig returns the modified kubeconfig with server url
func (c *Container) GetKubeConfig(ctx context.Context) ([]byte, error) {
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

// LoadImages loads images into the k3s container.
func (c *Container) LoadImages(ctx context.Context, images ...string) error {
	// save image
	imagesTar, err := os.CreateTemp(os.TempDir(), "images*.tar")
	if err != nil {
		return fmt.Errorf("creating temporary images file %w", err)
	}
	defer func() {
		_ = os.Remove(imagesTar.Name())
	}()

	err = tcimage.SaveImages(context.Background(), imagesTar.Name(), images...)
	if err != nil {
		return fmt.Errorf("saving images %w", err)
	}

	containerPath := fmt.Sprintf("/tmp/%s", filepath.Base(imagesTar.Name()))
	err = c.CopyFileToContainer(ctx, imagesTar.Name(), containerPath, 0x644)
	if err != nil {
		return fmt.Errorf("copying image to container %w", err)
	}

	_, _, err = c.Exec(ctx, []string{"ctr", "-n=k8s.io", "images", "import", containerPath})
	if err != nil {
		return fmt.Errorf("importing image %w", err)
	}

	return nil
}
