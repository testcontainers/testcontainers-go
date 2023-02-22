package pulsar

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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

type PulsarContainer struct {
	testcontainers.Container
	logConsumers []testcontainers.LogConsumer
	URI          string
}

type PulsarContainerRequest struct {
	testcontainers.ContainerRequest
	logConsumers []testcontainers.LogConsumer
}

// PulsarContainerOptions is a function that can be used to configure the Pulsar container
type PulsarContainerOptions func(req *PulsarContainerRequest)

// WithConfigModifier allows to override the default container config
func WithConfigModifier(modifier func(config *container.Config)) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.ConfigModifier = modifier
	}
}

// WithEndpointSettingsModifier allows to override the default endpoint settings
func WithEndpointSettingsModifier(modifier func(settings map[string]*network.EndpointSettings)) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.EnpointSettingsModifier = modifier
	}
}

// WithFunctionsWorker enables the functions worker, which will override the default pulsar command
// and add a waiting strategy for the functions worker
func WithFunctionsWorker() PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
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
func WithHostConfigModifier(modifier func(hostConfig *container.HostConfig)) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.HostConfigModifier = modifier
	}
}

// WithLogConsumer allows to add log consumers to the container
func WithLogConsumers(consumer ...testcontainers.LogConsumer) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.logConsumers = append(req.logConsumers, consumer...)
	}
}

// WithPulsarImage allows to override the default Pulsar image
func WithPulsarImage(image string) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		if image == "" {
			image = defaultPulsarImage
		}

		req.Image = image
	}
}

func WithTransactions() PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.ContainerRequest.Env["PULSAR_PREFIX_transactionCoordinatorEnabled"] = "true"

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
func StartContainer(ctx context.Context, opts ...PulsarContainerOptions) (*PulsarContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultPulsarImage,
		ExposedPorts: []string{defaultPulsarPort, defaultPulsarAdminPort},
		WaitingFor:   defaultWaitStrategies,
		Cmd:          []string{"/bin/bash", "-c", strings.Join([]string{defaultPulsarCmd, detaultPulsarCmdWithoutFunctionsWorker}, " ")},
	}

	pulsarRequest := PulsarContainerRequest{
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

	pulsarPort, err := c.MappedPort(ctx, defaultPulsarPort)
	if err != nil {
		return nil, err
	}

	pc := &PulsarContainer{
		Container:    c,
		logConsumers: pulsarRequest.logConsumers,
		URI:          fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
	}

	if len(pc.logConsumers) > 0 {
		c.StartLogProducer(ctx)
	}
	for _, lc := range pc.logConsumers {
		c.FollowOutput(lc)
	}

	return pc, nil
}
