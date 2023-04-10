package pulsar

import (
	"context"
	"fmt"
	"io"
	"strings"

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

// WithFunctionsWorker enables the functions worker, which will override the default pulsar command
// and add a waiting strategy for the functions worker
func WithFunctionsWorker() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = []string{"/bin/bash", "-c", defaultPulsarCmd}

		// add the waiting strategy for the functions worker
		defaultWaitStrategies.Strategies = append(
			defaultWaitStrategies.Strategies,
			wait.ForLog("Function worker service started"),
		)

		req.WaitingFor = defaultWaitStrategies
	}
}

// WithLogConsumers allows to add log consumers to the container.
// They will be automatically started and they will follow the container logs,
// but it's a responsibility of the caller to stop them calling StopLogProducer
func (c *Container) WithLogConsumers(ctx context.Context, consumer ...testcontainers.LogConsumer) {
	if len(c.LogConsumers) > 0 {
		c.StartLogProducer(ctx)
	}
	for _, lc := range c.LogConsumers {
		c.FollowOutput(lc)
	}
}

// WithPulsarEnv allows to use the native APIs and set each variable with PULSAR_PREFIX_ as prefix.
func WithPulsarEnv(configVar string, configValue string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["PULSAR_PREFIX_"+configVar] = configValue
	}
}

func WithTransactions() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
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

// RunContainer creates an instance of the Pulsar container type, being possible to pass a custom request and options
// The created container will use the following defaults:
// - image: docker.io/apachepulsar/pulsar:2.10.2
// - exposed ports: 6650/tcp, 8080/tcp
// - waiting strategy: wait for all the following strategies:
//		- the Pulsar admin API ("/admin/v2/clusters") to be ready on port 8080/tcp and return the response `["standalone"]`
// 		- the log message "Successfully updated the policies on namespace public/default"
// - command: "/bin/bash -c /pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss"
func RunContainer(ctx context.Context, opts ...testcontainers.CustomizeRequestOption) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultPulsarImage,
		Env:          map[string]string{},
		ExposedPorts: []string{defaultPulsarPort, defaultPulsarAdminPort},
		WaitingFor:   defaultWaitStrategies,
		Cmd:          []string{"/bin/bash", "-c", strings.Join([]string{defaultPulsarCmd, detaultPulsarCmdWithoutFunctionsWorker}, " ")},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt(&genericContainerReq)
	}

	c, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	pc := &Container{
		Container: c,
	}

	return pc, nil
}
