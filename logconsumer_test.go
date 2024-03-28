package testcontainers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const lastMessage = "DONE"

type TestLogConsumer struct {
	Msgs []string
	Done chan bool

	// Accepted provides a blocking way of ensuring the logs messages have been consumed.
	// This allows for proper synchronization during Test_StartStop in particular.
	// Please see func devNullAcceptorChan if you're not interested in this synchronization.
	Accepted chan string
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	s := string(l.Content)
	if s == fmt.Sprintf("echo %s\n", lastMessage) {
		g.Done <- true
		return
	}
	g.Accepted <- s
	g.Msgs = append(g.Msgs, s)
}

// devNullAcceptorChan returns string channel that essentially sends all strings to dev null
func devNullAcceptorChan() chan string {
	c := make(chan string)
	go func(c <-chan string) {
		for range c {
			// do nothing, just pull off channel
		}
	}(c)
	return c
}

func Test_LogConsumerGetsCalled(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&g},
		},
	}

	gReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := testcontainers.GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=hello")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=there")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	select {
	case <-g.Done:
	case <-time.After(5 * time.Second):
		t.Fatal("never received final log message")
	}

	assert.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs)

	terminateContainerOnEnd(t, ctx, c)
}

type TestLogTypeConsumer struct {
	LogTypes map[string]string
	Ack      chan bool
}

func (t *TestLogTypeConsumer) Accept(l testcontainers.Log) {
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		t.Ack <- true
		return
	}

	t.LogTypes[l.LogType] = string(l.Content)
}

func Test_ShouldRecognizeLogTypes(t *testing.T) {
	ctx := context.Background()

	g := TestLogTypeConsumer{
		LogTypes: map[string]string{},
		Ack:      make(chan bool),
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&g},
		},
	}

	gReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := testcontainers.GenericContainer(ctx, gReq)
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=this-is-stdout")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stderr?echo=this-is-stderr")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-g.Ack

	assert.Equal(t, map[string]string{
		testcontainers.StdoutLog: "echo this-is-stdout\n",
		testcontainers.StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
}

func Test_MultipleLogConsumers(t *testing.T) {
	ctx := context.Background()

	first := TestLogConsumer{
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
	}
	second := TestLogConsumer{
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&first, &second},
		},
	}

	gReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := testcontainers.GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-first.Done
	<-second.Done

	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, first.Msgs)
	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, second.Msgs)
	require.NoError(t, c.Terminate(ctx))
}

func TestContainerLogsShouldBeWithoutStreamHeader(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:      "alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	terminateContainerOnEnd(t, ctx, container)
	r, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "0", strings.TrimSpace(string(b)))
}

func TestContainerLogsEnableAtStart(t *testing.T) {
	ctx := context.Background()
	g := TestLogConsumer{
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
	}

	// logConsumersAtRequest {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&g},
		},
	}
	// }

	gReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := testcontainers.GenericContainer(ctx, gReq)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=hello")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=there")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	select {
	case <-g.Done:
	case <-time.After(10 * time.Second):
		t.Fatal("never received final log message")
	}
	assert.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs)

	terminateContainerOnEnd(t, ctx, c)
}
