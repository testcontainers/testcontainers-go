package testcontainers

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/log"
)

// Container is a container that can be created, started, stopped, and removed.
type Container interface {
	GetImage() string
	SessionID() string // get session id for the container
	Printf(format string, args ...interface{})
}

// CreatedContainer is a container that has been created.
type CreatedContainer interface {
	Container // embed the Container interface, as the created container is a container

	CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error
	CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error)
	CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error
	GetContainerID() string
	State(context.Context) (*types.ContainerState, error)
	Terminate(context.Context) error
}

type ReadyContainer interface {
	CreatedContainer // embed the CreatedContainer interface, as the started container is able to do everything a created container can do

	Host(context.Context) (string, error)
	Inspect(context.Context) (*types.ContainerJSON, error)
	MappedPort(context.Context, nat.Port) (nat.Port, error)
	PortEndpoint(context.Context, nat.Port, string) (string, error) // Alias for Host() + Port() methods
	ContainerIP(context.Context) (string, error)
	Logs(context.Context) (io.ReadCloser, error)
	Exec(context.Context, []string, ...exec.ProcessOption) (int, io.Reader, error)
	Stop(context.Context, *time.Duration) error

	// Readiness methods
	WaitUntilReady(ctx context.Context) error
}

// StartedContainer is a container that has been started.
type StartedContainer interface {
	ReadyContainer // embed the ReadyContainer interface, as the started container is a ready container

	// log consumer methods
	FollowOutput(lc log.Consumer) // TODO: do not expose this method, so all the hooks need a concrete Docker container
	StartLogProduction(ctx context.Context, opts ...log.ProductionOption) error
	StopLogProduction() error
	WithLogProductionTimeout(timeout time.Duration)
}
