package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/docker/docker/client"

	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

const lastMessage = "DONE"

type TestLogConsumer struct {
	Msgs []string
	Ack  chan bool
}

func (g *TestLogConsumer) Accept(l Log) {
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		g.Ack <- true
		return
	}

	g.Msgs = append(g.Msgs, string(l.Content))
}

func Test_LogConsumerGetsCalled(t *testing.T) {
	/*
		send one request at a time to a server that will
		print whatever was sent in the "echo" parameter, the log
		consumer should get all of the messages and append them
		to the Msgs slice
	*/

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
	if err != nil {
		t.Fatal(err)
	}

	ep, err := c.Endpoint(ctx, "http")
	if err != nil {
		t.Fatal(err)
	}

	g := TestLogConsumer{
		Msgs: []string{},
		Ack:  make(chan bool),
	}

	err = c.StartLogProducer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	c.FollowOutput(&g)

	_, err = http.Get(ep + "/stdout?echo=hello")
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(ep + "/stdout?echo=there")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	_, err = http.Get(ep + fmt.Sprintf("/stdout?echo=%s", lastMessage))
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-g.Ack:
	case <-time.After(5 * time.Second):
		t.Fatal("never received final log message")
	}
	_ = c.StopLogProducer()

	// get rid of the server "ready" log
	g.Msgs = g.Msgs[1:]

	assert.Equal(t, []string{"echo hello\n", "echo there\n"}, g.Msgs)
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
	if err != nil {
		t.Fatal(err)
	}
	terminateContainerOnEnd(t, ctx, c)

	ep, err := c.Endpoint(ctx, "http")
	if err != nil {
		t.Fatal(err)
	}

	g := TestLogTypeConsumer{
		LogTypes: map[string]string{},
		Ack:      make(chan bool),
	}

	err = c.StartLogProducer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	c.FollowOutput(&g)

	_, err = http.Get(ep + "/stdout?echo=this-is-stdout")
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(ep + "/stderr?echo=this-is-stderr")
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(ep + fmt.Sprintf("/stdout?echo=%s", lastMessage))
	if err != nil {
		t.Fatal(err)
	}

	<-g.Ack
	_ = c.StopLogProducer()

	assert.Equal(t, map[string]string{
		StdoutLog: "echo this-is-stdout\n",
		StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
}

func TestContainerLogWithErrClosed(t *testing.T) {
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

	client, err := testcontainersdocker.NewClient(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	provider := &DockerProvider{client: client, DockerProviderOptions: &DockerProviderOptions{GenericProviderOptions: &GenericProviderOptions{Logger: TestLogger(t)}}}

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

	var consumer TestLogConsumer
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
