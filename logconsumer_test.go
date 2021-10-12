package testcontainers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"
)

const lastMessage = "DONE"

type TestLogConsumer struct {
	Msgs []string
	Ack  chan bool
}

func (g *TestLogConsumer) Accept(l Log) {
	s := string(l.Content)
	if s == fmt.Sprintf("echo %s\n", lastMessage) {
		g.Ack <- true
		return
	}

	g.Msgs = append(g.Msgs, s)
}

func Test_LogConsumerGetsCalled(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	g := TestLogConsumer{
		Msgs: []string{},
		Ack:  make(chan bool),
	}

	c.FollowOutput(&g)

	err = c.StartLogProducer(ctx)
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=hello")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=there")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	select {
	case <-g.Ack:
	case <-time.After(5 * time.Second):
		t.Fatal("never received final log message")
	}
	assert.NilError(t, c.StopLogProducer())
	assert.DeepEqual(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs)
	assert.NilError(t, c.Terminate(ctx))
}

type TestLogTypeConsumer struct {
	LogTypes map[string]string
	Ack      chan bool
}

func (t *TestLogTypeConsumer) Accept(l Log) {
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		t.Ack <- true
		return
	}

	t.LogTypes[l.LogType] = string(l.Content)
}

func Test_ShouldRecognizeLogTypes(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	g := TestLogTypeConsumer{
		LogTypes: map[string]string{},
		Ack:      make(chan bool),
	}

	c.FollowOutput(&g)

	err = c.StartLogProducer(ctx)
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=this-is-stdout")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stderr?echo=this-is-stderr")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-g.Ack
	assert.NilError(t, c.StopLogProducer())

	assert.DeepEqual(t, map[string]string{
		StdoutLog: "echo this-is-stdout\n",
		StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
	assert.NilError(t, c.Terminate(ctx))
}

func Test_MultipleLogConsumers(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	first := TestLogConsumer{Msgs: []string{}, Ack: make(chan bool)}
	second := TestLogConsumer{Msgs: []string{}, Ack: make(chan bool)}

	c.FollowOutput(&first)
	c.FollowOutput(&second)

	err = c.StartLogProducer(ctx)
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-first.Ack
	<-second.Ack
	assert.NilError(t, c.StopLogProducer())

	assert.DeepEqual(t, []string{"ready\n", "echo mlem\n"}, first.Msgs)
	assert.DeepEqual(t, []string{"ready\n", "echo mlem\n"}, second.Msgs)
	assert.NilError(t, c.Terminate(ctx))
}

func Test_StartStop(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testresources/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	g := TestLogConsumer{Msgs: []string{}, Ack: make(chan bool)}

	c.FollowOutput(&g)

	require.NoError(t, c.StopLogProducer(), "nothing should happen even if the producer is not started")
	require.NoError(t, c.StartLogProducer(ctx))
	require.Error(t, c.StartLogProducer(ctx), "log producer is already started")

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	require.NoError(t, c.StopLogProducer())
	require.NoError(t, c.StartLogProducer(ctx))

	_, err = http.Get(ep + "/stdout?echo=mlem2")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-g.Ack
	// Do not close producer here, let's delegate it to c.Terminate

	assert.DeepEqual(t, []string{
		"ready\n",
		"echo mlem\n",
		"ready\n",
		"echo mlem\n",
		"echo mlem2\n",
	}, g.Msgs)
	assert.NilError(t, c.Terminate(ctx))
}
