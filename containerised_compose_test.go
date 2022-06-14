package testcontainers

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"io"
	"testing"
	"time"
)

type FakeContainerProvider struct {
}

func (p *FakeContainerProvider) CreateContainer(ctx context.Context, request ContainerRequest) (Container, error) {
	panic("not implemented")
}

func (p *FakeContainerProvider) Health(ctx context.Context) error {
	panic("not implemented")
}

type FakeContainer struct {
}

func (f FakeContainer) GetContainerID() string {
	panic("not implemented")
}

func (f FakeContainer) Endpoint(ctx context.Context, s string) (string, error) {
	panic("not implemented")
}

func (f FakeContainer) PortEndpoint(ctx context.Context, port nat.Port, s string) (string, error) {
	panic("not implemented")
}

func (f FakeContainer) Host(ctx context.Context) (string, error) {
	panic("not implemented")
}

func (f FakeContainer) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	panic("not implemented")
}

func (f FakeContainer) Ports(ctx context.Context) (nat.PortMap, error) {
	panic("not implemented")
}

func (f FakeContainer) SessionID() string {
	panic("not implemented")
}

func (f FakeContainer) Start(ctx context.Context) error {
	panic("not implemented")
}

func (f FakeContainer) Stop(ctx context.Context, duration *time.Duration) error {
	panic("not implemented")
}

func (f FakeContainer) Terminate(ctx context.Context) error {
	panic("not implemented")
}

func (f FakeContainer) Logs(ctx context.Context) (io.ReadCloser, error) {
	panic("not implemented")
}

func (f FakeContainer) FollowOutput(consumer LogConsumer) {
	panic("not implemented")
}

func (f FakeContainer) StartLogProducer(ctx context.Context) error {
	panic("not implemented")
}

func (f FakeContainer) StopLogProducer() error {
	panic("not implemented")
}

func (f FakeContainer) Name(ctx context.Context) (string, error) {
	panic("not implemented")
}

func (f FakeContainer) State(ctx context.Context) (*types.ContainerState, error) {
	return &types.ContainerState{
		Status:     "exited",
		FinishedAt: "2022-06-12",
	}, nil
}

func (f FakeContainer) Networks(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

func (f FakeContainer) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	panic("not implemented")
}

func (f FakeContainer) Exec(ctx context.Context, cmd []string) (int, io.Reader, error) {
	panic("not implemented")
}

func (f FakeContainer) ContainerIP(ctx context.Context) (string, error) {
	panic("not implemented")
}

func (f FakeContainer) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	panic("not implemented")
}

func (f FakeContainer) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	panic("not implemented")
}

func (f FakeContainer) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	panic("not implemented")
}

func (p *FakeContainerProvider) RunContainer(context.Context, ContainerRequest) (Container, error) {
	return FakeContainer{}, nil
}

func Test_NewContainerisedDockerCompose_UsingFakeContainer(t *testing.T) {
	ctx := context.Background()

	compose := NewContainerisedDockerCompose([]string{"docker-compose.yml"}, "test", ContainerisedDockerComposeOptions{
		Provider: &FakeContainerProvider{},
		Context:  ctx,
	})

	res := compose.Invoke()

	if res.Error != nil {
		t.Fatal()
	}
}

func Test_NewContainerisedDockerCompose_UsingDockerContainer(t *testing.T) {
	ctx := context.Background()

	compose := NewContainerisedDockerCompose([]string{"docker-compose.yml"}, "test", ContainerisedDockerComposeOptions{
		Provider: &DockerProvider{},
		Context:  ctx,
	})

	res := compose.Invoke()

	if res.Error != nil {
		t.Fatal()
	}
}
