package testcontainers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const lastMessage = "DONE"

type TestLogConsumer struct {
	mtx  sync.Mutex
	msgs []string
	Done chan struct{}

	// Accepted provides a blocking way of ensuring the logs messages have been consumed.
	// This allows for proper synchronization during TestStartStop in particular.
	// Please see func devNullAcceptorChan if you're not interested in this synchronization.
	Accepted chan string
}

func (g *TestLogConsumer) Accept(l log.Log) {
	s := string(l.Content)
	if s == fmt.Sprintf("echo %s\n", lastMessage) {
		close(g.Done)
		return
	}
	g.Accepted <- s

	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.msgs = append(g.msgs, s)
}

func (g *TestLogConsumer) Msgs() []string {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	return g.msgs
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

func TestLogConsumerGetsCalled(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&g},
		},
		Started: true,
	}

	c, err := New(ctx, req)
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

	assert.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs())

	TerminateContainerOnEnd(t, ctx, c)
}

type TestLogTypeConsumer struct {
	LogTypes map[string]string
	Ack      chan bool
}

func (t *TestLogTypeConsumer) Accept(l log.Log) {
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		t.Ack <- true
		return
	}

	t.LogTypes[l.LogType] = string(l.Content)
}

func TestShouldRecognizeLogTypes(t *testing.T) {
	ctx := context.Background()

	g := TestLogTypeConsumer{
		LogTypes: map[string]string{},
		Ack:      make(chan bool),
	}

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&g},
		},
		Started: true,
	}

	c, err := New(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)

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
		log.Stdout: "echo this-is-stdout\n",
		log.Stderr: "echo this-is-stderr\n",
	}, g.LogTypes)
}

func TestMultipleLogConsumers(t *testing.T) {
	ctx := context.Background()

	first := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}
	second := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&first, &second},
		},
		Started: true,
	}

	c, err := New(ctx, req)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-first.Done
	<-second.Done

	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, first.Msgs())
	assert.Equal(t, []string{"ready\n", "echo mlem\n"}, second.Msgs())
	require.NoError(t, c.Terminate(ctx))
}

func TestContainerLogWithErrClosed(t *testing.T) {
	if os.Getenv("GITHUB_RUN_ID") != "" {
		t.Skip("Skipping as flaky on GitHub Actions, Please see https://github.com/testcontainers/testcontainers-go/issues/1924")
	}

	t.Cleanup(func() {
		config.Reset()
	})

	// First spin up a docker-in-docker container, then spin up an inner container within that dind container
	// Logs are being read from the inner container via the dind container's tcp port, which can be briefly
	// closed to test behaviour in connection-closed situations.
	ctx := context.Background()

	dind, err := New(ctx, Request{
		Started:      true,
		Image:        "docker.io/docker:dind",
		ExposedPorts: []string{"2375/tcp"},
		Env:          map[string]string{"DOCKER_TLS_CERTDIR": ""},
		WaitingFor:   wait.ForListeningPort("2375/tcp"),
		Privileged:   true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, dind)

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

	dindCli, err := core.NewClient(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer dindCli.Close()

	// Inject the client into the context for the downstream API calls
	dindCtx := context.WithValue(ctx, ClientContextKey, dindCli)

	consumer := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	nginx, err := newContainer(dindCtx, Request{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&consumer},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := nginx.Start(dindCtx); err != nil {
		t.Fatal(err)
	}
	TerminateContainerOnEnd(t, dindCtx, nginx)

	port, err := nginx.MappedPort(dindCtx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}

	// Gather the initial container logs
	time.Sleep(time.Second * 1)
	existingLogs := len(consumer.Msgs())

	hitNginx := func() {
		i, _, err := dind.Exec(context.Background(), []string{"wget", "--spider", "localhost:" + port.Port()})
		if err != nil || i > 0 {
			t.Fatalf("Can't make request to nginx container from dind container")
		}
	}

	hitNginx()
	time.Sleep(time.Second * 1)
	msgs := consumer.Msgs()
	if len(msgs)-existingLogs != 1 {
		t.Fatalf("logConsumer should have 1 new log message, instead has: %v", msgs[existingLogs:])
	}
	existingLogs = len(consumer.Msgs())

	iptableArgs := []string{
		"INPUT", "-p", "tcp", "--dport", "2375",
		"-j", "REJECT", "--reject-with", "tcp-reset",
	}
	// Simulate a transient closed connection to the docker daemon
	i, _, err := dind.Exec(context.Background(), append([]string{"iptables", "-A"}, iptableArgs...))
	if err != nil || i > 0 {
		t.Fatalf("Failed to close connection to dind daemon: i(%d), err %v", i, err)
	}
	i, _, err = dind.Exec(context.Background(), append([]string{"iptables", "-D"}, iptableArgs...))
	if err != nil || i > 0 {
		t.Fatalf("Failed to re-open connection to dind daemon: i(%d), err %v", i, err)
	}
	time.Sleep(time.Second * 3)

	hitNginx()
	hitNginx()
	time.Sleep(time.Second * 1)
	msgs = consumer.Msgs()
	if len(msgs)-existingLogs != 2 {
		t.Fatalf(
			"LogConsumer should have 2 new log messages after detecting closed connection and"+
				" re-requesting logs. Instead has:\n%s", msgs[existingLogs:],
		)
	}
}

func TestContainerLogsShouldBeWithoutStreamHeader(t *testing.T) {
	ctx := context.Background()
	req := Request{
		Image:      "alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
		Started:    true,
	}
	container, err := New(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	TerminateContainerOnEnd(t, ctx, container)

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
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	// logConsumersAtRequest {
	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Opts:      []log.ProductionOption{log.WithProductionTimeout(10 * time.Second)},
			Consumers: []log.Consumer{&g},
		},
		Started: true,
	}
	// }

	c, err := New(ctx, req)
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
	assert.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs())

	TerminateContainerOnEnd(t, ctx, c)
}

func TestStartLogProductionStillStartsWithTooLowTimeout(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Opts:      []log.ProductionOption{log.WithProductionTimeout(4 * time.Second)},
			Consumers: []log.Consumer{&g},
		},
		Started: true,
	}

	c, err := New(ctx, req)
	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, c)
}

func TestStartLogProductionStillStartsWithTooHighTimeout(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Opts:      []log.ProductionOption{log.WithProductionTimeout(61 * time.Second)},
			Consumers: []log.Consumer{&g},
		},
		Started: true,
	}

	c, err := New(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, c)

	// because the log production timeout is too high, the container should have already been terminated
	// so no need to terminate it again with "testcontainers.testcontainers.TerminateContainerOnEnd(t, ctx, c)"
	require.NoError(t, c.StopLogProduction())

	TerminateContainerOnEnd(t, ctx, c)
}

func TestMultiContainerLogConsumer_CancelledContext(t *testing.T) {
	// Redirect stderr to a buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Context with cancellation functionality for simulating user interruption
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure it gets called.

	first := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	containerReq1 := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&first},
		},
		Started: true,
	}

	c, err := New(ctx, containerReq1)
	require.NoError(t, err)

	ep1, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep1 + "/stdout?echo=hello1")
	require.NoError(t, err)

	_, err = http.Get(ep1 + "/stdout?echo=there1")
	require.NoError(t, err)

	second := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	containerReq2 := Request{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &log.ConsumerConfig{
			Consumers: []log.Consumer{&second},
		},
		Started: true,
	}

	c2, err := New(ctx, containerReq2)
	require.NoError(t, err)

	ep2, err := c2.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep2 + "/stdout?echo=hello2")
	require.NoError(t, err)

	_, err = http.Get(ep2 + "/stdout?echo=there2")
	require.NoError(t, err)

	// Handling the termination of the containers
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer shutdownCancel()
		_ = c.Terminate(shutdownCtx)
		_ = c2.Terminate(shutdownCtx)
	}()

	// Deliberately calling context cancel
	cancel()

	// We check log size due to context cancellation causing
	// varying message counts, leading to test failure.
	assert.GreaterOrEqual(t, len(first.Msgs()), 2)
	assert.GreaterOrEqual(t, len(second.Msgs()), 2)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read the stderr output from the buffer
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Check the stderr message
	actual := buf.String()

	// The context cancel shouldn't cause the system to throw a
	// log.StoppedForOutOfSyncMessage, as it hangs the system with
	// the multiple containers.
	assert.False(t, strings.Contains(actual, log.StoppedForOutOfSyncMessage))
}

type FooLogConsumer struct {
	LogChannel chan string
}

func (c FooLogConsumer) Accept(rawLog Log) {
	log := string(rawLog.Content)
	c.LogChannel <- log
}

func NewFooLogConsumer() *FooLogConsumer {
	return &FooLogConsumer{
		LogChannel: make(chan string),
	}
}

func TestRestartContainerWithLogConsumer(t *testing.T) {
	logConsumer := NewFooLogConsumer()

	ctx := context.Background()
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:           "hello-world",
			AlwaysPullImage: true,
			LogConsumerCfg: &LogConsumerConfig{
				Consumers: []LogConsumer{logConsumer},
			},
		},
		Started: false,
	})
	if err != nil {
		t.Fatalf("Cant create container: %s", err.Error())
	}

	err = container.Start(ctx)
	if err != nil {
		t.Fatalf("Cant start container: %s", err.Error())
	}

	d := 30 * time.Second
	err = container.Stop(ctx, &d)
	if err != nil {
		t.Fatalf("Cant stop container: %s", err.Error())
	}
	err = container.Start(ctx)
	if err != nil {
		t.Fatalf("Cant start container: %s", err.Error())
	}

	for s := range logConsumer.LogChannel {
		if strings.Contains(s, "Hello from Docker!") {
			break
		}
	}
}
