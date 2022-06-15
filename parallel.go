package testcontainers

import (
	"context"
	"fmt"
	"sync"
)

const (
	defaultWorkersCount = 8
)

// GenericParallelOptions represents additional options for running
// * WorkersCount - count of parallel workers. if field empty(zero), default value will be 'defaultWorkersCount'
type GenericParallelOptions struct {
	WorkersCount int
}

// GenericContainerRequestError represents error from parallel request
type GenericContainerRequestError struct {
	Request GenericContainerRequest
	Error   error
}

type GenericParallelErrors struct {
	Errors []GenericContainerRequestError
}

func (gpe GenericParallelErrors) Error() string {
	return fmt.Sprintf("%v", gpe.Errors)
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
func GenericParallelContainers(ctx context.Context, reqs []GenericContainerRequest, opt GenericParallelOptions) ([]Container, error) {
	if opt.WorkersCount == 0 {
		opt.WorkersCount = defaultWorkersCount
	}

	tasksChanSize := opt.WorkersCount
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
