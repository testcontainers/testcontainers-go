package wait_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	hostAddress         = "0.0.0.0"
	defaultInternalPort = "80/tcp"
	defaultHostPort     = "60001"
)

// Mock arguments and responses.
var (
	anyStringSlice = mock.AnythingOfType("[]string")
	anyNatPort     = mock.AnythingOfType("nat.Port")
)

// containerState is an enum for the container state.
type containerState int

const (
	running containerState = iota
	dead
	oom
	exited
)

// waitBuilder is a builder for the wait tests.
type waitBuilder struct {
	mappedPortsResponses []nat.Port
	inspectResponse      *types.ContainerJSON
	stateResponse        *types.ContainerState
	logsResponse         io.ReadCloser
	equalError           string
	targetError          any
	hostPort             nat.Port
	internalPort         nat.Port
	execResponses        []int
}

// newWaitBuilder returns a fully initialised waitBuilder.
func newWaitBuilder() *waitBuilder {
	b := &waitBuilder{
		stateResponse: &types.ContainerState{Running: true},
		internalPort:  defaultInternalPort,
	}

	return b.MappedPorts("", defaultHostPort)
}

// SendingRequest sets the mapped ports response to the mapped port
// when true, otherwise the first response is a port not found error
// and the second response is the mapped port.
func (b *waitBuilder) SendingRequest(sendingRequest bool) *waitBuilder {
	if sendingRequest {
		return b.MappedPorts(b.hostPort)
	}

	return b.MappedPorts("", b.hostPort)
}

// InternalPort sets the internal port used for generating the port map
// of inspect responses.
func (b *waitBuilder) InternalPort(port nat.Port) *waitBuilder {
	b.internalPort = port
	return b
}

// MappedPorts sets the mapped ports responses.
func (b *waitBuilder) MappedPorts(hostPorts ...nat.Port) *waitBuilder {
	b.mappedPortsResponses = hostPorts
	if len(hostPorts) == 0 {
		b.hostPort = ""
		b.inspectResponse.NetworkSettings.Ports = nat.PortMap{}
		return b
	}

	b.hostPort = hostPorts[len(hostPorts)-1]

	if b.hostPort != "" {
		return b.InspectPort(b.hostPort)
	}

	return b
}

// MappedPort sets the mapped port.
func (b *waitBuilder) MappedPort(hostPort nat.Port) *waitBuilder {
	b.hostPort = hostPort
	if len(b.mappedPortsResponses) == 0 {
		b.mappedPortsResponses = append(b.mappedPortsResponses, hostPort)
	} else {
		b.mappedPortsResponses[len(b.mappedPortsResponses)-1] = hostPort
	}

	return b.InspectPort(hostPort)
}

// InspectPort sets inspect response for the given port using the
// default internal port.
func (b *waitBuilder) InspectPort(port nat.Port) *waitBuilder {
	return b.InspectPortMap(nat.PortMap{
		b.internalPort: []nat.PortBinding{{
			HostIP:   hostAddress,
			HostPort: port.Port(),
		}},
	})
}

// InspectPortMap sets inspect response for the given port map.
func (b *waitBuilder) InspectPortMap(ports nat.PortMap) *waitBuilder {
	b.inspectResponse = &types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				NetworkMode: network.NetworkDefault,
			},
		},
		NetworkSettings: &types.NetworkSettings{
			NetworkSettingsBase: types.NetworkSettingsBase{
				Ports: ports,
			},
		},
	}

	return b
}

// State sets the state response.
func (b *waitBuilder) State(state containerState) *waitBuilder {
	switch state {
	case running:
		b.stateResponse = &types.ContainerState{Running: true}
		b.equalError = ""
	case dead:
		b.stateResponse = &types.ContainerState{Status: "dead"}
		b.equalError = `unexpected container status "dead"`
	case oom:
		b.stateResponse = &types.ContainerState{OOMKilled: true}
		b.equalError = "container crashed with out-of-memory (OOMKilled)"
	case exited:
		b.stateResponse = &types.ContainerState{Status: "exited", ExitCode: 1}
		b.equalError = "container exited with code 1"
	default:
		panic(fmt.Sprintf("unknown state: %d", state))
	}
	return b
}

// EqualError sets the the expected error, if blank expects no error.
func (b *waitBuilder) EqualError(expected string) *waitBuilder {
	b.equalError = expected
	b.targetError = nil
	return b
}

// ErrorAs sets the the expected error, if blank expects no error.
func (b *waitBuilder) ErrorAs(target any) *waitBuilder {
	b.equalError = ""
	b.targetError = target
	return b
}

// Exec configures the strategy to expect the given exec response.
func (b *waitBuilder) Exec(responses ...int) *waitBuilder {
	b.execResponses = responses
	return b
}

// Logs configures the strategy to expect the given logs response.
func (b *waitBuilder) Logs(r io.Reader) *waitBuilder {
	b.logsResponse = io.NopCloser(r)
	return b
}

// Ports returns the number of ports in the inspect response.
func (b *waitBuilder) Ports() int {
	return len(b.inspectResponse.NetworkSettings.Ports)
}

// NoTCP returns true if the inspect response has no tcp ports, false otherwise.
func (b *waitBuilder) NoTCP() bool {
	for port := range b.inspectResponse.NetworkSettings.Ports {
		if port.Proto() == "tcp" {
			return false
		}
	}

	return true
}

// Target builds and returns the mockStrategyTarget with the given settings.
func (b *waitBuilder) Target() *mockStrategyTarget {
	target := &mockStrategyTarget{}
	targetExpect := target.EXPECT()
	targetExpect.Host(anyContext).Return("localhost", nil)
	if len(b.mappedPortsResponses) > 0 {
		resp := targetExpect.MappedPort(anyContext, anyNatPort)
		for _, port := range b.mappedPortsResponses {
			if port == "" {
				resp = resp.Return("", wait.PortNotFoundErr(b.hostPort))
			} else {
				resp = resp.Return(port, nil)
			}
		}
	}
	targetExpect.State(anyContext).Return(b.stateResponse, nil)
	targetExpect.Inspect(anyContext).Return(b.inspectResponse, nil)
	if len(b.execResponses) > 0 {
		resp := targetExpect.Exec(anyContext, anyStringSlice)
		for _, response := range b.execResponses {
			resp = resp.Return(response, nil, nil)
		}
	}
	if b.logsResponse != nil {
		targetExpect.Logs(anyContext).Return(b.logsResponse, nil)
	}

	return target
}

// Run runs the validation with the given strategy.
func (b *waitBuilder) Run(t *testing.T, strategy wait.Strategy) {
	t.Helper()

	b.RunTarget(t, b.Target(), strategy)
}

// RunTarget runs the validation with the given target and strategy.
func (b *waitBuilder) RunTarget(t *testing.T, target *mockStrategyTarget, strategy wait.Strategy) {
	t.Helper()

	err := strategy.WaitUntilReady(context.Background(), target)
	if b.equalError != "" {
		require.EqualError(t, err, b.equalError)
		return
	}

	if b.targetError != nil {
		require.ErrorAs(t, err, b.targetError)
		return
	}

	require.NoError(t, err)
}
