package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RegistryContainer represents the Registry container type used in the module
type RegistryContainer struct {
	testcontainers.Container
	RegistryName string
}

// Address returns the address of the Registry container, using the HTTP protocol
func (c *RegistryContainer) Address(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, "5000")
	if err != nil {
		return "", err
	}

	ipAddress, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", ipAddress, port.Port()), nil
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

	pushOpts := types.ImagePushOptions{
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

// RunContainer creates an instance of the Registry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*RegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "registry:2.8.3",
		ExposedPorts: []string{"5000/tcp"},
		Env: map[string]string{
			// convenient for testing
			"REGISTRY_STORAGE_DELETE_ENABLED": "true",
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("listening on [::]:5000").WithStartupTimeout(10*time.Second),
		),
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
	if err != nil {
		return nil, err
	}

	c := &RegistryContainer{Container: container}

	address, err := c.Address(ctx)
	if err != nil {
		return c, err
	}

	c.RegistryName = strings.TrimPrefix(address, "http://")

	return c, nil
}
