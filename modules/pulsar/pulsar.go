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

const (
	defaultPulsarPort                      = "6650/tcp"
	defaultPulsarAdminPort                 = "8080/tcp"
	defaultPulsarCmd                       = "/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone"
	detaultPulsarCmdWithoutFunctionsWorker = "--no-functions-worker -nss"
	transactionTopicEndpoint               = "/admin/v2/persistent/pulsar/system/transaction_coordinator_assign/partitions"
)

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
	LogConsumers []testcontainers.LogConsumer // Deprecated. Use the ContainerRequest instead. Needs to be exported to control the stop from the caller
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
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = []string{"/bin/bash", "-c", defaultPulsarCmd}

		ss := []wait.Strategy{
			wait.ForLog("Function worker service started"),
		}

		ss = append(ss, defaultWaitStrategies.Strategies...)

		req.WaitingFor = wait.ForAll(ss...)

		return nil
	}
}

// Deprecated: use the testcontainers.WithLogConsumers functional option instead
// WithLogConsumers allows to add log consumers to the container.
// They will be automatically started and they will follow the container logs,
// but it's a responsibility of the caller to stop them calling StopLogProducer
func (c *Container) WithLogConsumers(ctx context.Context, consumer ...testcontainers.LogConsumer) {
	if len(c.LogConsumers) > 0 {
		// not handling the error because it will return an error if and only if the producer is already started
		_ = c.StartLogProducer(ctx)
	}
	for _, lc := range c.LogConsumers {
		c.FollowOutput(lc)
	}
}

// WithPulsarEnv allows to use the native APIs and set each variable with PULSAR_PREFIX_ as prefix.
func WithPulsarEnv(configVar string, configValue string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["PULSAR_PREFIX_"+configVar] = configValue

		return nil
	}
}

func WithTransactions() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if err := WithPulsarEnv("transactionCoordinatorEnabled", "true")(req); err != nil {
			return err
		}

		// clone defaultWaitStrategies
		ss := []wait.Strategy{
			wait.ForHTTP(transactionTopicEndpoint).WithPort(defaultPulsarAdminPort).WithStatusCodeMatcher(func(statusCode int) bool {
				return statusCode == 200
			}),
		}

		ss = append(ss, defaultWaitStrategies.Strategies...)

		req.WaitingFor = wait.ForAll(ss...)

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Pulsar container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	return Run(ctx, "docker.io/apachepulsar/pulsar:2.10.2", opts...)
}

// Run creates an instance of the Pulsar container type, being possible to pass a custom request and options
// The created container will use the following defaults:
// - image: docker.io/apachepulsar/pulsar:2.10.2
// - exposed ports: 6650/tcp, 8080/tcp
// - waiting strategy: wait for all the following strategies:
//   - the Pulsar admin API ("/admin/v2/clusters") to be ready on port 8080/tcp and return the response `["standalone"]`
//   - the log message "Successfully updated the policies on namespace public/default"
//
// - command: "/bin/bash -c /pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss"
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
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
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
