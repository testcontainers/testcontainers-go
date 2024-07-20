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
	"github.com/testcontainers/testcontainers-go/wait"
)

const lastMessage = "DONE"

type TestLogConsumer struct {
	mtx  sync.Mutex
	msgs []string
	Done chan struct{}

	// Accepted provides a blocking way of ensuring the logs messages have been consumed.
	// This allows for proper synchronization during Test_StartStop in particular.
	// Please see func devNullAcceptorChan if you're not interested in this synchronization.
	Accepted chan string
}

func (g *TestLogConsumer) Accept(l Log) {
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

func Test_LogConsumerGetsCalled(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&g},
		},
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
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

	require.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs())
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

	g := TestLogTypeConsumer{
		LogTypes: map[string]string{},
		Ack:      make(chan bool),
	}

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&g},
		},
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
	require.NoError(t, err)

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
		StdoutLog: "echo this-is-stdout\n",
		StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
}

func Test_MultipleLogConsumers(t *testing.T) {
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

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&first, &second},
		},
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
	require.NoError(t, err)

	ep, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=mlem")
	require.NoError(t, err)

	_, err = http.Get(ep + "/stdout?echo=" + lastMessage)
	require.NoError(t, err)

	<-first.Done
	<-second.Done

	expected := []string{"ready\n", "echo mlem\n"}
	require.Equal(t, expected, first.Msgs())
	require.Equal(t, expected, second.Msgs())
}

func TestContainerLogWithErrClosed(t *testing.T) {
	if os.Getenv("GITHUB_RUN_ID") != "" {
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
	CleanupContainer(t, dind)
	require.NoError(t, err)

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
		time.Sleep(10 * time.Millisecond)
		t.Log("retrying get endpoint")
	}
	if err != nil {
		t.Fatal("get endpoint:", err)
	}

	opts := []client.Opt{client.WithHost(remoteDocker), client.WithAPIVersionNegotiation()}

	dockerClient, err := NewDockerClientWithOpts(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	provider := &DockerProvider{
		client: dockerClient,
		config: config.Read(),
		DockerProviderOptions: &DockerProviderOptions{
			GenericProviderOptions: &GenericProviderOptions{
				Logger: TestLogger(t),
			},
		},
	}

	consumer := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	nginx, err := provider.CreateContainer(ctx, ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&consumer},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := nginx.Start(ctx); err != nil {
		t.Fatal(err)
	}
	CleanupContainer(t, nginx)

	port, err := nginx.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}

	// Gather the initial container logs
	time.Sleep(time.Second * 1)
	existingLogs := len(consumer.Msgs())

	hitNginx := func() {
		i, _, err := dind.Exec(ctx, []string{"wget", "--spider", "localhost:" + port.Port()})
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
	i, _, err := dind.Exec(ctx, append([]string{"iptables", "-A"}, iptableArgs...))
	if err != nil || i > 0 {
		t.Fatalf("Failed to close connection to dind daemon: i(%d), err %v", i, err)
	}
	i, _, err = dind.Exec(ctx, append([]string{"iptables", "-D"}, iptableArgs...))
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
	req := ContainerRequest{
		Image:      "alpine:latest",
		Cmd:        []string{"sh", "-c", "id -u"},
		WaitingFor: wait.ForExit(),
	}
	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	CleanupContainer(t, ctr)
	if err != nil {
		t.Fatal(err)
	}

	r, err := ctr.Logs(ctx)
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
	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Opts:      []LogProductionOption{WithLogProductionTimeout(10 * time.Second)},
			Consumers: []LogConsumer{&g},
		},
	}
	// }

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
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
	require.Equal(t, []string{"ready\n", "echo hello\n", "echo there\n"}, g.Msgs())
}

func Test_StartLogProductionStillStartsWithTooLowTimeout(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Opts:      []LogProductionOption{WithLogProductionTimeout(4 * time.Second)},
			Consumers: []LogConsumer{&g},
		},
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
	require.NoError(t, err)
}

func Test_StartLogProductionStillStartsWithTooHighTimeout(t *testing.T) {
	ctx := context.Background()

	g := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Opts:      []LogProductionOption{WithLogProductionTimeout(61 * time.Second)},
			Consumers: []LogConsumer{&g},
		},
	}

	gReq := GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	c, err := GenericContainer(ctx, gReq)
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	dc := c.(*DockerContainer)
	require.NoError(t, dc.stopLogProduction())
}

func Test_MultiContainerLogConsumer_CancelledContext(t *testing.T) {
	// Redirect stderr to a buffer
	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() {
		// Restore stderr
		os.Stderr = oldStderr
		w.Close()
	}()

	// Context with cancellation functionality for simulating user interruption
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure it gets called.

	first := TestLogConsumer{
		msgs:     []string{},
		Done:     make(chan struct{}),
		Accepted: devNullAcceptorChan(),
	}

	containerReq1 := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&first},
		},
	}

	genericReq1 := GenericContainerRequest{
		ContainerRequest: containerReq1,
		Started:          true,
	}

	c, err := GenericContainer(ctx, genericReq1)
	CleanupContainer(t, c)
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

	containerReq2 := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "./testdata/",
			Dockerfile: "echoserver.Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("ready"),
		LogConsumerCfg: &LogConsumerConfig{
			Consumers: []LogConsumer{&second},
		},
	}

	genericReq2 := GenericContainerRequest{
		ContainerRequest: containerReq2,
		Started:          true,
	}

	c2, err := GenericContainer(ctx, genericReq2)
	CleanupContainer(t, c2)
	require.NoError(t, err)

	ep2, err := c2.Endpoint(ctx, "http")
	require.NoError(t, err)

	_, err = http.Get(ep2 + "/stdout?echo=hello2")
	require.NoError(t, err)

	_, err = http.Get(ep2 + "/stdout?echo=there2")
	require.NoError(t, err)

	// Deliberately calling context cancel
	cancel()

	// We check log size due to context cancellation causing
	// varying message counts, leading to test failure.
	assert.GreaterOrEqual(t, len(first.Msgs()), 2)
	assert.GreaterOrEqual(t, len(second.Msgs()), 2)

	// Close the pipe so as not to block on empty.
	w.Close()

	// Read the stderr output from the buffer
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Check the stderr message
	actual := buf.String()

	// The context cancel shouldn't cause the system to throw a
	// logStoppedForOutOfSyncMessage, as it hangs the system with
	// the multiple containers.
	require.NotContains(t, actual, logStoppedForOutOfSyncMessage)
}

// FooLogConsumer is a test log consumer that accepts logs from the
// "hello-world" Docker image, which prints out the "Hello from Docker!"
// log message.
type FooLogConsumer struct {
	LogChannel chan string
	t          *testing.T
}

// Accept receives a log message and sends it to the log channel if it
// contains the "Hello from Docker!" message.
func (c FooLogConsumer) Accept(rawLog Log) {
	log := string(rawLog.Content)
	if strings.Contains(log, "Hello from Docker!") {
		select {
		case c.LogChannel <- log:
		default:
		}
	}
}

// AssertRead waits for a log message to be received.
func (c FooLogConsumer) AssertRead() {
	select {
	case <-c.LogChannel:
	case <-time.After(5 * time.Second):
		c.t.Fatal("receive timeout")
	}
}

// SlurpOne reads a value from the channel if it is available.
func (c FooLogConsumer) SlurpOne() {
	select {
	case <-c.LogChannel:
	default:
	}
}

func NewFooLogConsumer(t *testing.T) *FooLogConsumer {
	t.Helper()

	return &FooLogConsumer{
		t:          t,
		LogChannel: make(chan string, 2),
	}
}

func TestRestartContainerWithLogConsumer(t *testing.T) {
	logConsumer := NewFooLogConsumer(t)

	ctx := context.Background()
	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:           "hello-world",
			AlwaysPullImage: true,
			LogConsumerCfg: &LogConsumerConfig{
				Consumers: []LogConsumer{logConsumer},
			},
		},
		Started: false,
	})
	CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Start and confirm that the log consumer receives the log message.
	err = ctr.Start(ctx)
	require.NoError(t, err)

	logConsumer.AssertRead()

	// Stop the container and clear any pending message.
	d := 5 * time.Second
	err = ctr.Stop(ctx, &d)
	require.NoError(t, err)

	logConsumer.SlurpOne()

	// Restart the container and confirm that the log consumer receives new log messages.
	err = ctr.Start(ctx)
	require.NoError(t, err)

	// First message is from the first start.
	logConsumer.AssertRead()
	logConsumer.AssertRead()
}
