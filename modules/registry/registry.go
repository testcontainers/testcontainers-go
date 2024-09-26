package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// registryPort is the default port used by the Registry container.
	registryPort = "5000/tcp"

	// DefaultImage is the default image used by the Registry container.
	DefaultImage = "registry:2.8.3"
)

// RegistryContainer represents the Registry container type used in the module
type RegistryContainer struct {
	testcontainers.Container
	RegistryName string
}

// Address returns the address of the Registry container, using the HTTP protocol
func (c *RegistryContainer) Address(ctx context.Context) (string, error) {
	host, err := c.HostAddress(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + host, nil
}

// HostAddress returns the host address including port of the Registry container.
func (c *RegistryContainer) HostAddress(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, registryPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	host, err := c.Container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	if host == "localhost" {
		// This is a workaround for WSL, where localhost is not reachable from Docker.
		host, err = localAddress(ctx)
		if err != nil {
			return "", fmt.Errorf("local ip: %w", err)
		}
	}

	return net.JoinHostPort(host, port.Port()), nil
}

// localAddress returns the local address of the machine
// which can be used to connect to the local registry.
// This avoids the issues with localhost on WSL.
func localAddress(ctx context.Context) (string, error) {
	if os.Getenv("WSL_DISTRO_NAME") == "" {
		return "localhost", nil
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", "golang.org:80")
	if err != nil {
		return "", fmt.Errorf("dial: %w", err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// getEndpointWithAuth returns the HTTP endpoint of the Registry container, along with the image auth
// for the image referece.
// E.g. imageRef = "localhost:5000/alpine:latest"
func getEndpointWithAuth(ctx context.Context, imageRef string) (string, string, registry.AuthConfig, error) {
	registry, imageAuth, err := testcontainers.DockerImageAuth(ctx, imageRef)
	if err != nil {
		return "", "", imageAuth, fmt.Errorf("failed to get image auth: %w", err)
	}

	imageWithoutRegistry := strings.TrimPrefix(imageRef, registry+"/")
	image := strings.Split(imageWithoutRegistry, ":")[0]
	tag := strings.Split(imageWithoutRegistry, ":")[1]

	return fmt.Sprintf("/v2/%s/manifests/%s", image, tag), image, imageAuth, nil
}

// DeleteImage deletes an image reference from the Registry container.
// It will use the HTTP endpoint of the Registry container to delete it,
// doing a HEAD request to get the image digest and then a DELETE request
// to actually delete the image.
// E.g. imageRef = "localhost:5000/alpine:latest"
func (c *RegistryContainer) DeleteImage(ctx context.Context, imageRef string) error {
	endpoint, image, imageAuth, err := getEndpointWithAuth(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to get image auth: %w", err)
	}

	var digest string
	err = wait.ForHTTP(endpoint).
		WithMethod(http.MethodHead).
		WithBasicAuth(imageAuth.Username, imageAuth.Password).
		WithHeaders(map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json"}).
		WithStatusCodeMatcher(func(statusCode int) bool {
			return statusCode == http.StatusOK
		}).
		WithResponseHeadersMatcher(func(headers http.Header) bool {
			contentDigest := headers.Get("Docker-Content-Digest")
			if contentDigest != "" {
				digest = contentDigest
				return true
			}

			return false
		}).
		WaitUntilReady(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get image digest: %w", err)
	}

	deleteEndpoint := fmt.Sprintf("/v2/%s/manifests/%s", image, digest)
	return wait.ForHTTP(deleteEndpoint).
		WithMethod(http.MethodDelete).
		WithBasicAuth(imageAuth.Username, imageAuth.Password).
		WithStatusCodeMatcher(func(statusCode int) bool {
			return statusCode == http.StatusAccepted
		}).
		WaitUntilReady(ctx, c)
}

// ImageExists checks if an image exists in the Registry container. It will use the v2 HTTP endpoint
// of the Registry container to check if the image reference exists.
// E.g. imageRef = "localhost:5000/alpine:latest"
func (c *RegistryContainer) ImageExists(ctx context.Context, imageRef string) error {
	endpoint, _, imageAuth, err := getEndpointWithAuth(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to get image auth: %w", err)
	}

	return wait.ForHTTP(endpoint).
		WithMethod(http.MethodHead).
		WithBasicAuth(imageAuth.Username, imageAuth.Password).
		WithHeaders(map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json"}).
		WithForcedIPv4LocalHost().
		WithStatusCodeMatcher(func(statusCode int) bool {
			return statusCode == http.StatusOK
		}).
		WithResponseHeadersMatcher(func(headers http.Header) bool {
			return headers.Get("Docker-Content-Digest") != ""
		}).
		WaitUntilReady(ctx, c)
}

// PushImage pushes an image to the Registry container. It will use the internally stored RegistryName
// to push the image to the container, and it will finally wait for the image to be pushed.
func (c *RegistryContainer) PushImage(ctx context.Context, ref string) error {
	dockerProvider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return fmt.Errorf("failed to create Docker provider: %w", err)
	}
	defer dockerProvider.Close()

	dockerCli := dockerProvider.Client()

	_, imageAuth, err := testcontainers.DockerImageAuth(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to get image auth: %w", err)
	}

	pushOpts := image.PushOptions{
		All: true,
	}

	// see https://github.com/docker/docs/blob/e8e1204f914767128814dca0ea008644709c117f/engine/api/sdk/examples.md?plain=1#L649-L657
	encodedJSON, err := json.Marshal(imageAuth)
	if err != nil {
		return fmt.Errorf("failed to encode image auth: %w", err)
	} else {
		pushOpts.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	}

	_, err = dockerCli.ImagePush(ctx, ref, pushOpts)
	if err != nil {
		return fmt.Errorf("failed to push image %s: %w", ref, err)
	}

	return c.ImageExists(ctx, ref)
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Registry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error) {
	return Run(ctx, "registry:2.8.3", opts...)
}

// Run creates an instance of the Registry container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{registryPort},
		Env: map[string]string{
			// convenient for testing
			"REGISTRY_STORAGE_DELETE_ENABLED": "true",
		},
		WaitingFor: wait.ForHTTP("/").
			WithPort(registryPort).
			WithStartupTimeout(10 * time.Second),
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
	var c *RegistryContainer
	if container != nil {
		c = &RegistryContainer{Container: container}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	address, err := c.Address(ctx)
	if err != nil {
		return c, fmt.Errorf("address: %w", err)
	}

	c.RegistryName = strings.TrimPrefix(address, "http://")

	return c, nil
}

// SetDockerAuthConfig sets the DOCKER_AUTH_CONFIG environment variable with
// authentication for the given host, username and password sets.
// It returns a function to reset the environment back to the previous state.
func SetDockerAuthConfig(host, username, password string, additional ...string) (func(), error) {
	authConfigs, err := DockerAuthConfig(host, username, password, additional...)
	if err != nil {
		return nil, fmt.Errorf("docker auth config: %w", err)
	}

	auth, err := json.Marshal(dockercfg.Config{AuthConfigs: authConfigs})
	if err != nil {
		return nil, fmt.Errorf("marshal auth config: %w", err)
	}

	previousAuthConfig := os.Getenv("DOCKER_AUTH_CONFIG")
	os.Setenv("DOCKER_AUTH_CONFIG", string(auth))

	return func() {
		if previousAuthConfig == "" {
			os.Unsetenv("DOCKER_AUTH_CONFIG")
			return
		}
		os.Setenv("DOCKER_AUTH_CONFIG", previousAuthConfig)
	}, nil
}

// DockerAuthConfig returns a map of AuthConfigs including base64 encoded Auth field
// for the provided details. It also accepts additional host, username and password
// triples to add more auth configurations.
func DockerAuthConfig(host, username, password string, additional ...string) (map[string]dockercfg.AuthConfig, error) {
	if len(additional)%3 != 0 {
		return nil, fmt.Errorf("additional must be a multiple of 3")
	}

	additional = append(additional, host, username, password)
	authConfigs := make(map[string]dockercfg.AuthConfig, len(additional)/3)
	for i := 0; i < len(additional); i += 3 {
		host, username, password := additional[i], additional[i+1], additional[i+2]
		auth := dockercfg.AuthConfig{
			Username: username,
			Password: password,
		}

		if username != "" || password != "" {
			auth.Auth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		}
		authConfigs[host] = auth
	}

	return authConfigs, nil
}
