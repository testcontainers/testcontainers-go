package localstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/mod/semver"
)

const defaultPort = 4566
const defaultRegion = "us-east-1"
const defaultVersion = "1.4.0"
const hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"
const localstackHostEnvVar = "LOCALSTACK_HOST"

const defaultAccessKeyID = "accesskey"
const defaultSecretAccessKey = "secretkey"
const defaultToken = "token"

func isLegacyMode(image string) bool {
	parts := strings.Split(image, ":")
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

// StartContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func StartContainer(ctx context.Context, overrideReq OverrideContainerRequestOption) (*LocalStackContainer, error) {
	// defaultContainerRequest {
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("localstack/localstack:%s", defaultVersion),
		Binds:        []string{fmt.Sprintf("%s:/var/run/docker.sock", testcontainersdocker.ExtractDockerHost(ctx))},
		WaitingFor:   wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStartupTimeout(120 * time.Second),
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", defaultPort)},
		Env:          map[string]string{},
	}
	// }

	localStackReq := LocalStackContainerRequest{
		ContainerRequest: req,
	}

	// first, when needed, we merge the user request with the default one
	if overrideReq != nil {
		merged := overrideReq(localStackReq.ContainerRequest)
		localStackReq.ContainerRequest = merged
	}

	if isLegacyMode(localStackReq.Image) {
		return nil, fmt.Errorf("version=%s. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0", localStackReq.Image)
	}

	if isVersion2(localStackReq.Image) {
		hostnameExternalReason, err := configureDockerHost(&localStackReq, localstackHostEnvVar)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Setting %s to %s (%s)\n", localstackHostEnvVar, req.Env[localstackHostEnvVar], hostnameExternalReason)
	} else {
		hostnameExternalReason, err := configureDockerHost(&localStackReq, hostnameExternalEnvVar)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Setting %s to %s (%s)\n", hostnameExternalEnvVar, req.Env[hostnameExternalEnvVar], hostnameExternalReason)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: localStackReq.ContainerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	c := &LocalStackContainer{
		Container: container,
	}
	return c, nil
}

func configureDockerHost(req *LocalStackContainerRequest, envVar string) (reason string, err error) {
	err = nil
	reason = ""

	if _, ok := req.Env[envVar]; ok {
		reason = "explicitly as environment variable"
		return
	}

	// if the container is not connected to the default network, use the last network alias in the first network
	// for that we need to check if the container is connected to a network and if it has network aliases
	if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
		alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

		req.Env[envVar] = alias
		reason = "to match last network alias on container with non-default network"
		return
	}

	var dockerProvider *testcontainers.DockerProvider
	dockerProvider, err = testcontainers.NewDockerProvider()
	if err != nil {
		return
	}
	defer dockerProvider.Close()

	var daemonHost string
	daemonHost, err = dockerProvider.DaemonHost(context.Background())
	if err != nil {
		return
	}

	req.Env[envVar] = daemonHost
	reason = "to match host-routable address for container"

	return
}
