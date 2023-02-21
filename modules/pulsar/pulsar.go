package pulsar

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultPulsarImage = "docker.io/apachepulsar/pulsar:2.10.2"
const defaultPulsarPort = "6650/tcp"
const defaultPulsarAdminPort = "8080/tcp"

type PulsarContainer struct {
	testcontainers.Container
	URI string
}

type PulsarContainerRequest struct {
	testcontainers.ContainerRequest
}

// PulsarContainerOptions is a function that can be used to configure the Pulsar container
type PulsarContainerOptions func(req *PulsarContainerRequest)

// WithCmd allows to override the default command for the container
func WithCmd(cmd []string) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.Cmd = cmd
	}
}

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

// WithHostConfigModifier allows to override the default host config
func WithHostConfigModifier(modifier func(hostConfig *container.HostConfig)) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.HostConfigModifier = modifier
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

// WithWaitingFor allows to override the default waiting strategy
func WithWaitingFor(waitingFor wait.Strategy) PulsarContainerOptions {
	return func(req *PulsarContainerRequest) {
		req.WaitingFor = waitingFor
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
	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}
	req := testcontainers.ContainerRequest{
		Image:        defaultPulsarImage,
		ExposedPorts: []string{defaultPulsarPort, defaultPulsarAdminPort},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/admin/v2/clusters").WithPort(defaultPulsarAdminPort).WithResponseMatcher(matchAdminResponse),
			wait.ForLog("Successfully updated the policies on namespace public/default"),
		),
		Cmd: []string{
			"/bin/bash",
			"-c",
			"/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss",
		},
	}

	pulsarRequest := PulsarContainerRequest{
		ContainerRequest: req,
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

	c.StartLogProducer(ctx)
	defer c.StopLogProducer()
	lc := logConsumer{}
	c.FollowOutput(&lc)

	pulsarPort, err := c.MappedPort(ctx, defaultPulsarPort)
	if err != nil {
		return nil, err
	}

	return &PulsarContainer{
		Container: c,
		URI:       fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
	}, nil
}

type logConsumer struct{}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
