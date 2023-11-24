package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
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

func (g *TestLogConsumer) Accept(l Log) {
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
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
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
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
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
	case <-g.Done:
	case <-time.After(5 * time.Second):
		t.Fatal("never received final log message")
	}
	assert.Nil(t, c.StopLogProducer())
	assert.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs)

	terminateContainerOnEnd(t, ctx, c)
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
			Context:    "./testdata/",
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
	terminateContainerOnEnd(t, ctx, c)

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
	assert.Nil(t, c.StopLogProducer())

	assert.Equal(t, map[string]string{
		StdoutLog: "echo this-is-stdout\n",
		StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
}

func Test_MultipleLogConsumers(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
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

	c.FollowOutput(&first)
	c.FollowOutput(&second)

	err = c.StartLogProducer(ctx)
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-first.Done
	<-second.Done
	assert.Nil(t, c.StopLogProducer())

	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, first.Msgs)
	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, second.Msgs)
	assert.Nil(t, c.Terminate(ctx))
}

func Test_StartStop(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
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
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: make(chan string),
	}
	c.FollowOutput(&g)

	require.NoError(t, c.StopLogProducer(), "nothing should happen even if the producer is not started")

	require.NoError(t, c.StartLogProducer(ctx))
	require.Equal(t, <-g.Accepted, "ready\n")

	require.Error(t, c.StartLogProducer(ctx), "log producer is already started")

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)
	require.Equal(t, <-g.Accepted, "echo mlem\n")

	require.NoError(t, c.StopLogProducer())

	require.NoError(t, c.StartLogProducer(ctx))
	require.Equal(t, <-g.Accepted, "ready\n")
	require.Equal(t, <-g.Accepted, "echo mlem\n")

	_, err = http.Get(ep + "/stdout?echo=mlem2")
	require.NoError(t, err)
	require.Equal(t, <-g.Accepted, "echo mlem2\n")

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-g.Done
	// Do not close producer here, let's delegate it to c.Terminate

	assert.Equal(t, []string{
		"ready\n",
		"echo mlem\n",
		"ready\n",
		"echo mlem\n",
		"echo mlem2\n",
	}, g.Msgs)
	assert.Nil(t, c.Terminate(ctx))
}

func TestContainerLogWithErrClosed(t *testing.T) {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		t.Skip("Skipping as flaky on GitHub Actions, Please see https://github.com/testcontainers/testcontainers-go/issues/1924")
	}

	t.Cleanup(func() {
		config.Reset()
	})

	if providerType == ProviderPodman {
		t.Skip("Docker-in-Docker does not work with rootless Podman")
	}
	// First spin up a docker-in-docker container, then spin up an inner container within that dind container
	// Logs are being read from the inner container via the dind container's tcp port, which can be briefly
	// closed to test behaviour in connection-closed situations.
	ctx := context.Background()

	dind, err := GenericContainer(ctx, GenericContainerRequest{
		Started: true,
		ContainerRequest: ContainerRequest{
			Image:        "docker.io/docker:dind",
			ExposedPorts: []string{"2375/tcp"},
			Env:          map[string]string{"DOCKER_TLS_CERTDIR": ""},
			WaitingFor:   wait.ForListeningPort("2375/tcp"),
			Privileged:   true,
		},
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, dind)

	var remoteDocker string

	ctxEndpoint, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// todo: remove this temporary fix (test is flaky).
	for {
		remoteDocker, err = dind.Endpoint(ctxEndpoint, "2375/tcp")
		if err == nil {
			break
		}
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}
		time.Sleep(100 * time.Microsecond)
		t.Log("retrying get endpoint")
	}
	if err != nil {
		t.Fatal("get endpoint:", err)
	}

	opts := []client.Opt{client.WithHost(remoteDocker), client.WithAPIVersionNegotiation()}

	client, err := NewDockerClientWithOpts(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	provider := &DockerProvider{
		client: client,
		config: ReadConfig(),
		DockerProviderOptions: &DockerProviderOptions{
			GenericProviderOptions: &GenericProviderOptions{
				Logger: TestLogger(t),
			},
		},
	}

	nginx, err := provider.CreateContainer(ctx, ContainerRequest{Image: "nginx", ExposedPorts: []string{"80/tcp"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := nginx.Start(ctx); err != nil {
		t.Fatal(err)
	}
	terminateContainerOnEnd(t, ctx, nginx)

	port, err := nginx.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}

	consumer := TestLogConsumer{
		Msgs:     []string{},
		Done:     make(chan bool),
		Accepted: devNullAcceptorChan(),
	}

	if err = nginx.StartLogProducer(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = nginx.StopLogProducer()
	}()
	nginx.FollowOutput(&consumer)

	// Gather the initial container logs
	time.Sleep(time.Second * 1)
	existingLogs := len(consumer.Msgs)

	hitNginx := func() {
		i, _, err := dind.Exec(ctx, []string{"wget", "--spider", "localhost:" + port.Port()})
		if err != nil || i > 0 {
			t.Fatalf("Can't make request to nginx container from dind container")
		}
	}

	hitNginx()
	time.Sleep(time.Second * 1)
	if len(consumer.Msgs)-existingLogs != 1 {
		t.Fatalf("logConsumer should have 1 new log message, instead has: %v", consumer.Msgs[existingLogs:])
	}
	existingLogs = len(consumer.Msgs)

	iptableArgs := []string{
		"INPUT", "-p", "tcp", "--dport", "2375",
		"-j", "REJECT", "--reject-with", "tcp-reset",
	}
	// Simulate a transient closed connection to the docker daemon
	i, _, err := dind.Exec(ctx, append([]string{"iptables", "-A"}, iptableArgs...))
	if err != nil || i > 0 {
		t.Fatalf("Failed to close connection to dind daemon")
	}
	i, _, err = dind.Exec(ctx, append([]string{"iptables", "-D"}, iptableArgs...))
	if err != nil || i > 0 {
		t.Fatalf("Failed to re-open connection to dind daemon")
	}
	time.Sleep(time.Second * 3)

	hitNginx()
	hitNginx()
	time.Sleep(time.Second * 1)
	if len(consumer.Msgs)-existingLogs != 2 {
		t.Fatalf(
			"LogConsumer should have 2 new log messages after detecting closed connection and"+
				" re-requesting logs. Instead has:\n%s", consumer.Msgs[existingLogs:],
		)
	}
}

func TestContainerLogsShouldBeWithoutStreamHeader(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image:      "alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
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
