package localstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/mod/semver"
)

const defaultPort = 4566
const defaultVersion = "0.11.2"
const hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"

// localStackContainer represents the LocalStack container type used in the module
type localStackContainer struct {
	testcontainers.Container
	legacyMode bool
}

func runInLegacyMode(version string) bool {
	if version == "latest" {
		return false
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v0.11") < 0
	}

	fmt.Printf("Version %s is not a semantic version, LocalStack will run in legacy mode.\n", version)
	fmt.Printf("Consider using \"setupLocalStack(context context.Context, version string, legacyMode bool)\" constructor if you want to disable legacy mode.")
	return true
}

// setupLocalStack creates an instance of the LocalStack container type
func setupLocalStack(ctx context.Context, version string, legacyMode bool, opts ...localStackContainerOption) (*localStackContainer, error) {
	if version == "" {
		version = defaultVersion
	}

	/*
		Do not run in legacy mode when the version is a valid semver version greater than the v0.11 and legacyMode is false
			| runInLegacyMode | legacyMode | result |
			|-----------------|------------|--------|
			| false           | false      | false  |
			| false           | true       | true   |
			| true            | false      | true   |
			| true            | true       | true   |
	*/
	legacyMode = !(!runInLegacyMode(version) && !legacyMode)

	req := testcontainers.ContainerRequest{
		Image:      "localstack/localstack:0.11.2",
		Binds:      []string{fmt.Sprintf("%s:/var/run/docker.sock", testcontainersdocker.ExtractDockerHost(ctx))},
		WaitingFor: wait.ForLog("Ready.\n").WithOccurrence(1),
	}

	localStackReq := LocalStackContainerRequest{
		ContainerRequest: req,
		legacyMode:       legacyMode,
		version:          version,
	}

	for _, opt := range opts {
		opt(&localStackReq)
	}

	hostnameExternalReason, err := configure(&localStackReq)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Setting %s to %s (%s)\n", hostnameExternalEnvVar, req.Env[hostnameExternalEnvVar], hostnameExternalReason)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &localStackContainer{Container: container, legacyMode: legacyMode}, nil
}

func configure(req *LocalStackContainerRequest) (reason string, err error) {
	err = nil
	reason = ""

	if _, ok := req.Env[hostnameExternalEnvVar]; ok {
		reason = "explicitly as environment variable"
		return
	}

	// if the container is not connected to the default network, use the last network alias in the first network
	// for that we need to check if the container is connected to a network and if it has network aliases
	if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
		alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

		req.Env[hostnameExternalEnvVar] = alias
		reason = "to match last network alias on container with non-default network"
		return
	}

	var dockerProvider *testcontainers.DockerProvider
	dockerProvider, err = testcontainers.NewDockerProvider()
	if err != nil {
		return
	}

	var daemonHost string
	daemonHost, err = dockerProvider.DaemonHost(context.Background())
	if err != nil {
		return
	}

	req.Env[hostnameExternalEnvVar] = daemonHost
	reason = "to match host-routable address for container"

	return
}
