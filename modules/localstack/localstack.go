package localstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort            = 4566
	defaultVersion         = "1.4.0"
	hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"
	localstackHostEnvVar   = "LOCALSTACK_HOST"
)

func isLegacyMode(img string) bool {
	parts := strings.Split(img, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return false
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v0.11") < 0 // version < v0.11
	}

	return true
}

func isVersion2(image string) bool {
	parts := strings.Split(image, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return true
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v2.0") > 0 // version >= v2.0
	}

	return true
}

// RunContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func RunContainer(ctx context.Context, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	dockerHost := testcontainers.ExtractDockerSocket()

	req := testcontainers.Request{
		Image:        fmt.Sprintf("localstack/localstack:%s", defaultVersion),
		WaitingFor:   wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStartupTimeout(120 * time.Second),
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", defaultPort)},
		Env:          map[string]string{},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{fmt.Sprintf("%s:/var/run/docker.sock", dockerHost)}
		},
		Started: true,
		Logger:  tclog.StandardLogger(),
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	if isLegacyMode(req.Image) {
		return nil, fmt.Errorf("version=%s. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0", req.Image)
	}

	envVar := hostnameExternalEnvVar
	if isVersion2(req.Image) {
		envVar = localstackHostEnvVar
	}

	hostnameExternalReason, err := configureDockerHost(&req, envVar)
	if err != nil {
		return nil, err
	}
	req.Logger.Printf("Setting %s to %s (%s)\n", envVar, req.Env[envVar], hostnameExternalReason)

	ctr, err := testcontainers.New(ctx, req)
	if err != nil {
		return nil, err
	}

	c := &Container{
		DockerContainer: ctr,
	}
	return c, nil
}

func configureDockerHost(req *testcontainers.Request, envVar string) (string, error) {
	reason := ""

	if _, ok := req.Env[envVar]; ok {
		return "explicitly as environment variable", nil
	}

	// if the container is not connected to the default network, use the last network alias in the first network
	// for that we need to check if the container is connected to a network and if it has network aliases
	if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
		alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

		req.Env[envVar] = alias
		return "to match last network alias on container with non-default network", nil
	}

	daemonHost, err := testcontainers.DaemonHost(context.Background())
	if err != nil {
		return reason, err
	}

	req.Env[envVar] = daemonHost
	return "to match host-routable address for container", nil
}
