package testcontainers

import (
	"context"
	"fmt"
	"sync"
)

const (
	MaxGenericWorkers = 8
)

type GenericParallelErrors struct {
	Errors []GenericContainerRequestError
}

func (gpe GenericParallelErrors) Error() string {
	return fmt.Sprintf("%v", gpe.Errors)
}

// GenericContainerRequestError represents error from parallel request
type GenericContainerRequestError struct {
	Request GenericContainerRequest
	Error   error
}

// GenericContainerRequest represents parameters to a generic container
type GenericContainerRequest struct {
	ContainerRequest              // embedded request for provider
	Started          bool         // whether to auto-start the container
	ProviderType     ProviderType // which provider to use, Docker if empty
	Logger           Logging      // provide a container specific Logging - use default global logger if empty
}

// GenericNetworkRequest represents parameters to a generic network
type GenericNetworkRequest struct {
	NetworkRequest              // embedded request for provider
	ProviderType   ProviderType // which provider to use, Docker if empty
}

// GenericNetwork creates a generic network with parameters
func GenericNetwork(ctx context.Context, req GenericNetworkRequest) (Network, error) {
	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}
	network, err := provider.CreateNetwork(ctx, req.NetworkRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create network", err)
	}

	return network, nil
}

func genericContainerRunner(
	ctx context.Context,
	requests <-chan GenericContainerRequest,
	errors chan<- GenericContainerRequestError,
	containers chan<- Container,
	wg *sync.WaitGroup) {
	for req := range requests {
		c, err := GenericContainer(ctx, req)
		if err != nil {
			errors <- GenericContainerRequestError{
				Request: req,
				Error:   err,
			}
			continue
		}
		containers <- c
	}
	wg.Done()
}

// GenericParallelContainers creates a generic containers with parameters in parallel mode
func GenericParallelContainers(ctx context.Context, reqs []GenericContainerRequest) ([]Container, error) {
	tasksChanSize := MaxGenericWorkers
	if tasksChanSize > len(reqs) {
		tasksChanSize = len(reqs)
	}

	tasksChan := make(chan GenericContainerRequest, tasksChanSize)
	errsChan := make(chan GenericContainerRequestError)
	resChan := make(chan Container)
	waitRes := make(chan struct{})

	containers := make([]Container, 0)
	errors := make([]GenericContainerRequestError, 0)

	wg := sync.WaitGroup{}
	wg.Add(tasksChanSize)

	// run workers
	for i := 0; i < tasksChanSize; i++ {
		go genericContainerRunner(ctx, tasksChan, errsChan, resChan, &wg)
	}

	go func() {
		for {
			select {
			case c, ok := <-resChan:
				if !ok {
					resChan = nil
				} else {
					containers = append(containers, c)
				}
			case e, ok := <-errsChan:
				if !ok {
					errsChan = nil
				} else {
					errors = append(errors, e)
				}
			}

			if resChan == nil && errsChan == nil {
				waitRes <- struct{}{}
				break
			}
		}

	}()

	for _, req := range reqs {
		tasksChan <- req
	}
	close(tasksChan)
	wg.Wait()
	close(resChan)
	close(errsChan)

	<-waitRes

	if len(errors) == 0 {
		return containers, GenericParallelErrors{Errors: errors}
	}

	return containers, nil

}

// GenericContainer creates a generic container with parameters
func GenericContainer(ctx context.Context, req GenericContainerRequest) (Container, error) {
	logging := req.Logger
	if logging == nil {
		logging = Logger
	}
	provider, err := req.ProviderType.GetProvider(WithLogger(logging))
	if err != nil {
		return nil, err
	}

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create container", err)
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return c, fmt.Errorf("%w: failed to start container", err)
		}
	}

	return c, nil
}

// GenericProvider represents an abstraction for container and network providers
type GenericProvider interface {
	ContainerProvider
	NetworkProvider
}
