package k3s

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/yaml.v3"
	//kubeCon "github.com/giantswarm/kubeconfig/"
)

var (
	defaultKubeSecurePort     = "6443/tcp"
	defaultRancherWebhookPort = "8443/tcp"
	// kubeConfigYaml            string
)

// K3sContainer represents the K3s container type used in the module
type K3sContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the K3s container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*K3sContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "docker.io/rancher/k3s:latest",
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

func WithNetwork(networks []string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Networks = networks
	}
}

func WithNetworkAlias(networkAlias map[string][]string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.NetworkAliases = networkAlias
	}
}

func WithKubectlConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: filepath.Join(".", "config"),
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)
	}
}

func WithCmd(commands []string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = commands
	}
}

// getkubeConfigYaml returns the string of modifed kubeconfig with server url
func (c *K3sContainer) getkubeConfigYaml(ctx context.Context) (string, error) {
	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get hostIP: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, nat.Port(defaultKubeSecurePort))
	if err != nil {
		return "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	reader, err := c.CopyFileFromContainer(ctx, "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to copy file from container: %w", err)
	}

	kubeConfigYaml, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file from container: %w", err)
	}

	server := "https://" + fmt.Sprintf("%v:%d", hostIP, mappedPort.Int())
	newkubeConfigYaml, err := kubeConfigYamlwithServer(string(kubeConfigYaml), server)
	if err != nil {
		return "", fmt.Errorf("failed to modify kubeconfig with server url: %w", err)
	}

	fmt.Println(newkubeConfigYaml)

	return newkubeConfigYaml, nil
}

func kubeConfigYamlwithServer(kubeConfigYaml, server string) (string, error) {

	kubeConfig, err := unmarshal([]byte(kubeConfigYaml))
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal kubeconfig: %w", err)
	}

	kubeConfig.Clusters[0].Cluster.Server = server
	modifiedKubeConfig, err := marshal(kubeConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	modifiedKubeConfigYaml := string(modifiedKubeConfig)
	return modifiedKubeConfigYaml, nil
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

//TODO
// getInternalKubeConfigYaml returns the string of modifed kubeconfig with server
// func (c *K3sContainer) getInternalKubeConfigYaml(ctx context.Context, networkAlias string) (string, error) {
// 	time.Sleep(time.Minute * 1)
// 	reader, err := c.CopyFileFromContainer(ctx, "/etc/rancher/k3s/k3s.yaml")
// 	if err != nil {
// 		return "", fmt.Errorf("failed to copy file from container: %w", err)
// 	}

// 	kubeConfigYaml, err := io.ReadAll(reader)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read file from container: %w", err)
// 	}
// 	//	time.Sleep(time.Minute * 1)
// 	nA, err := c.Container.NetworkAliases(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get hostIP: %w", err)
// 	}

// 	fmt.Println(nA)

// 	if _, ok := nA[networkAlias]; ok {
// 		server := "https://" + fmt.Sprintf("%s:%s", networkAlias, defaultKubeSecurePort)
// 		newkubeConfigYaml, err := kubeConfigYamlwithServer(string(kubeConfigYaml), server)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to read file from container: %w", err)
// 		}
// 		return newkubeConfigYaml, nil
// 	} else {
// 		return "", fmt.Errorf("%s is not a network alias for k3s container", networkAlias)
// 	}
// }
