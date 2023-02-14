package pulsar

import (
	"context"
	"fmt"
	"io"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type pulsarContainer struct {
	testcontainers.Container
	URI string
}

func startContainer(ctx context.Context) (*pulsarContainer, error) {
	matchAdminResponse := func(r io.Reader) bool {
		respBytes, _ := io.ReadAll(r)
		resp := string(respBytes)
		return resp == `["standalone"]`
	}
	pulsarRequest := testcontainers.ContainerRequest{
		Image:        "docker.io/apachepulsar/pulsar:2.10.2",
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/admin/v2/clusters").WithPort("8080/tcp").WithResponseMatcher(matchAdminResponse),
			wait.ForLog("Successfully updated the policies on namespace public/default"),
		),
		Cmd: []string{
			"/bin/bash",
			"-c",
			"/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss",
		},
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

	pulsarPort, err := c.MappedPort(ctx, "6650/tcp")
	if err != nil {
		return nil, err
	}

	return &pulsarContainer{
		Container: c,
		URI:       fmt.Sprintf("pulsar://127.0.0.1:%v", pulsarPort.Int()),
	}, nil
}

type logConsumer struct{}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}
