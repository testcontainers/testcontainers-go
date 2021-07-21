package testcontainers

import (
	"context"
	"fmt"
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
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		g.Ack <- true
		return
	}

	g.Msgs = append(g.Msgs, string(l.Content))
}

func Test_LogConsumerGetsCalled(t *testing.T) {
	t.Skip("This test is randomly failing for different versions of Go")
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
	c.StopLogProducer()

	// get rid of the server "ready" log
	g.Msgs = g.Msgs[1:]

	assert.DeepEqual(t, []string{"echo hello\n", "echo there\n"}, g.Msgs)
	c.Terminate(ctx)
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
	c.StopLogProducer()

	assert.DeepEqual(t, map[string]string{
		StdoutLog: "echo this-is-stdout\n",
		StderrLog: "echo this-is-stderr\n",
	}, g.LogTypes)
	c.Terminate(ctx)

}
