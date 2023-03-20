package pulsar

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultPulsarImage = "docker.io/apachepulsar/pulsar:2.10.2"
const defaultPulsarPort = "6650/tcp"
const defaultPulsarAdminPort = "8080/tcp"
const defaultPulsarCmd = "/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone"
const detaultPulsarCmdWithoutFunctionsWorker = "--no-functions-worker -nss"
const transactionTopicEndpoint = "/admin/v2/persistent/pulsar/system/transaction_coordinator_assign/partitions"

var defaultWaitStrategies = wait.ForAll(
	wait.ForHTTP("/admin/v2/clusters").WithPort(defaultPulsarAdminPort).WithResponseMatcher(func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}),
	wait.ForLog("Successfully updated the policies on namespace public/default"),
)

type Container struct {
	testcontainers.Container
	LogConsumers []testcontainers.LogConsumer // Needs to be exported to control the stop from the caller
}

func (c *Container) BrokerURL(ctx context.Context) (string, error) {
	return c.resolveURL(ctx, defaultPulsarPort)
}

func (c *Container) HTTPServiceURL(ctx context.Context) (string, error) {
	return c.resolveURL(ctx, defaultPulsarAdminPort)
}

func (c *Container) resolveURL(ctx context.Context, port nat.Port) (string, error) {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return "", err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return "", err
	}

	pulsarPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	proto := "pulsar"
	if port == defaultPulsarAdminPort {
		proto = "http"
	}

	return fmt.Sprintf("%s://%s:%v", proto, host, pulsarPort.Int()), nil
}

type ContainerRequest struct {
	testcontainers.ContainerRequest
	logConsumers []testcontainers.LogConsumer
}

// ContainerOptions is a function that can be used to configure the Pulsar container
type ContainerOptions func(req *ContainerRequest)

// WithConfigModifier allows to override the default container config
func WithConfigModifier(modifier func(config *container.Config)) ContainerOptions {
	return func(req *ContainerRequest) {
		req.ConfigModifier = modifier
	}
}

// WithEndpointSettingsModifier allows to override the default endpoint settings
func WithEndpointSettingsModifier(modifier func(settings map[string]*network.EndpointSettings)) ContainerOptions {
	return func(req *ContainerRequest) {
		req.EnpointSettingsModifier = modifier
	}
}

// WithFunctionsWorker enables the functions worker, which will override the default pulsar command
// and add a waiting strategy for the functions worker
func WithFunctionsWorker() ContainerOptions {
	return func(req *ContainerRequest) {
		req.Cmd = []string{"/bin/bash", "-c", defaultPulsarCmd}

		// add the waiting strategy for the functions worker
		defaultWaitStrategies.Strategies = append(
			defaultWaitStrategies.Strategies,
			wait.ForLog("Function worker service started"),
		)

		req.WaitingFor = defaultWaitStrategies
	}
}

// WithHostConfigModifier allows to override the default host config
func WithHostConfigModifier(modifier func(hostConfig *container.HostConfig)) ContainerOptions {
	return func(req *ContainerRequest) {
		req.HostConfigModifier = modifier
	}
}

// WithLogConsumer allows to add log consumers to the container. They will be automatically started and stopped by the StartContainer function
// but it's a responsibility of the caller to stop them calling StopLogProducer
func WithLogConsumers(consumer ...testcontainers.LogConsumer) ContainerOptions {
	return func(req *ContainerRequest) {
		req.logConsumers = append(req.logConsumers, consumer...)
	}
}

// WithPulsarEnv allows to use the native APIs and set each variable with PULSAR_PREFIX_ as prefix.
func WithPulsarEnv(configVar string, configValue string) ContainerOptions {
	return func(req *ContainerRequest) {
		req.ContainerRequest.Env["PULSAR_PREFIX_"+configVar] = configValue
	}
}

// WithPulsarImage allows to override the default Pulsar image
func WithPulsarImage(image string) ContainerOptions {
	return func(req *ContainerRequest) {
		if image == "" {
			image = defaultPulsarImage
		}

		req.Image = image
	}
}

func WithTransactions() ContainerOptions {
	return func(req *ContainerRequest) {
		WithPulsarEnv("transactionCoordinatorEnabled", "true")(req)

		// add the waiting strategy for the transaction topic
		defaultWaitStrategies.Strategies = append(
			defaultWaitStrategies.Strategies,
			wait.ForHTTP(transactionTopicEndpoint).WithPort(defaultPulsarAdminPort).WithStatusCodeMatcher(func(statusCode int) bool {
				return statusCode == 200
			}),
		)

		req.WaitingFor = defaultWaitStrategies
	}
}

// StartContainer creates an instance of the Pulsar container type, being possible to pass a custom request and options
// The created container will use the following defaults:
// - image: docker.io/apachepulsar/pulsar:2.10.2
// - exposed ports: 6650/tcp, 8080/tcp
// - waiting strategy: wait for all the following strategies:
//		- the Pulsar admin API ("/admin/v2/clusters") to be ready on port 8080/tcp and return the response `["standalone"]`
// 		- the log message "Successfully updated the policies on namespace public/default"
// - command: "/bin/bash -c /pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss"
func StartContainer(ctx context.Context, opts ...ContainerOptions) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultPulsarImage,
		Env:          map[string]string{},
		ExposedPorts: []string{defaultPulsarPort, defaultPulsarAdminPort},
		WaitingFor:   defaultWaitStrategies,
		Cmd:          []string{"/bin/bash", "-c", strings.Join([]string{defaultPulsarCmd, detaultPulsarCmdWithoutFunctionsWorker}, " ")},
	}

	pulsarRequest := ContainerRequest{
		ContainerRequest: req,
		logConsumers:     []testcontainers.LogConsumer{},
	}

	for _, opt := range opts {
		opt(&pulsarRequest)
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pulsarRequest.ContainerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	pc := &Container{
		Container:    c,
		LogConsumers: pulsarRequest.logConsumers,
	}

	if len(pc.LogConsumers) > 0 {
		c.StartLogProducer(ctx)
	}
	for _, lc := range pc.LogConsumers {
		c.FollowOutput(lc)
	}

	return pc, nil
}
