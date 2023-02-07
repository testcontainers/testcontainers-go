package localstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/mod/semver"
)

const defaultPort = 4566
const defaultRegion = "us-east-1"
const defaultVersion = "0.11.2"
const hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"

const defaultAccessKeyID = "accesskey"
const defaultSecretAccessKey = "secretkey"
const defaultToken = "token"

// LocalStackContainer represents the LocalStack container type used in the module
type LocalStackContainer struct {
	testcontainers.Container
	Credentials     Credentials
	Region          string
	EnabledServices map[string]Service
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Token           string
}

// ServicePort returns the port of the given service
func (l *LocalStackContainer) ServicePort(ctx context.Context, service EnabledService) (nat.Port, error) {
	if _, ok := l.EnabledServices[service.Name()]; !ok {
		return "", fmt.Errorf("service %s is not enabled", service.Name())
	}

	internalPort := l.EnabledServices[service.Name()].servicePort()

	p, err := nat.NewPort("tcp", fmt.Sprintf("%d", internalPort))
	if err != nil {
		return "", err
	}

	return l.MappedPort(ctx, p)
}

func runInLegacyMode(version string) bool {
	if version == "latest" {
		return false
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v0.11") < 0 // version < v0.11
	}

	fmt.Printf("Version %s is not a semantic version, LocalStack will run in legacy mode.\n", version)
	fmt.Printf("Consider using \"StartContainer(ctx, NoopOverrideContainerRequest, WithLegacyMode)\" constructor if you want to enable legacy mode.")
	return true
}

// StartContainer creates an instance of the LocalStack container type
func StartContainer(ctx context.Context, overrideReq overrideContainerRequestOption, opts ...localStackContainerOption) (*LocalStackContainer, error) {
	credentials := Credentials{
		AccessKeyID:     defaultAccessKeyID,
		SecretAccessKey: defaultSecretAccessKey,
		Token:           defaultToken,
	}

	req := testcontainers.ContainerRequest{
		Image:      "localstack/localstack",
		Binds:      []string{fmt.Sprintf("%s:/var/run/docker.sock", testcontainersdocker.ExtractDockerHost(ctx))},
		WaitingFor: wait.ForLog("Ready.\n").WithOccurrence(1).WithStartupTimeout(2 * time.Minute),
		Env: map[string]string{
			"AWS_ACCESS_KEY_ID":     credentials.AccessKeyID,
			"AWS_SECRET_ACCESS_KEY": credentials.SecretAccessKey,
			"AWS_SESSION_TOKEN":     credentials.Token,
		},
	}

	localStackReq := LocalStackContainerRequest{
		ContainerRequest: req,
	}

	for _, opt := range opts {
		opt(&localStackReq)
	}

	if localStackReq.version == "" {
		WithDefaultVersion(&localStackReq)
	}
	// use the passed version as image tag
	localStackReq.Image = fmt.Sprintf("%s:%s", localStackReq.Image, localStackReq.version)

	/*
		Do not run in legacy mode when the version is a valid semver version greater than the v0.11 and legacyMode is false
			| runInLegacyMode | legacyMode | result |
			|-----------------|------------|--------|
			| false           | false      | false  |
			| false           | true       | true   |
			| true            | false      | true   |
			| true            | true       | true   |
	*/
	localStackReq.legacyMode = !(!runInLegacyMode(localStackReq.version) && !localStackReq.legacyMode)

	if localStackReq.region == "" {
		WithDefaultRegion(&localStackReq)
	}

	hostnameExternalReason, err := configure(&localStackReq)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Setting %s to %s (%s)\n", hostnameExternalEnvVar, req.Env[hostnameExternalEnvVar], hostnameExternalReason)

	// at the end, when needed, we merge the user request with the default one
	if overrideReq != nil {
		merged := overrideReq(localStackReq.ContainerRequest)
		localStackReq.ContainerRequest = merged
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: localStackReq.ContainerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	enabledServices := make(map[string]Service)
	for _, service := range localStackReq.enabledServices {
		enabledServices[service.Name()] = service
	}

	c := &LocalStackContainer{
		Container:       container,
		Credentials:     credentials,
		EnabledServices: enabledServices,
		Region:          localStackReq.region,
	}
	return c, nil
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
