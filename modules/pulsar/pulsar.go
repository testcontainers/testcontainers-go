package pulsar

import (
	"context"
	"fmt"
	"io"

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

// PulsarContainerOptions is a function that can be used to configure the Pulsar container
type PulsarContainerOptions func(req *testcontainers.ContainerRequest)

// WithCmd allows to override the default command for the container
func WithCmd(cmd []string) PulsarContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		req.Cmd = cmd
	}
}

// WithEnv will merge the given environment variables with the default ones
func WithEnv(env map[string]string) PulsarContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		for k, v := range env {
			req.Env[k] = v
		}
	}
}

// WithPulsarImage allows to override the default Pulsar image
func WithPulsarImage(image string) PulsarContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		if image == "" {
			image = defaultPulsarImage
		}

		req.Image = image
	}
}

// WithWaitingFor allows to override the default waiting strategy
func WithWaitingFor(waitingFor wait.Strategy) PulsarContainerOptions {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = waitingFor
	}
}

func StartContainer(ctx context.Context, opts ...PulsarContainerOptions) (*PulsarContainer, error) {
	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}
	pulsarRequest := testcontainers.ContainerRequest{
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

	for _, opt := range opts {
		opt(&pulsarRequest)
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pulsarRequest,
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
