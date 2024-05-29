package testcontainers

import (
	"context"
	"fmt"
	"sync"
)

const (
	defaultWorkersCount = 8
)

type ParallelRequest []Request

// ParallelContainersOptions represents additional options for parallel running
type ParallelContainersOptions struct {
	WorkersCount int // count of parallel workers. If field empty(zero), default value will be 'defaultWorkersCount'
}

// ParallelContainersRequestError represents error from parallel request
type ParallelContainersRequestError struct {
	Request Request
	Error   error
}

type ParallelContainersError struct {
	Errors []ParallelContainersRequestError
}

func (gpe ParallelContainersError) Error() string {
	return fmt.Sprintf("%v", gpe.Errors)
}

func parallelContainersRunner(
	ctx context.Context,
	requests <-chan Request,
	errors chan<- ParallelContainersRequestError,
	containers chan<- StartedContainer,
	wg *sync.WaitGroup,
) {
	for req := range requests {
		c, err := New(ctx, req)
		if err != nil {
			errors <- ParallelContainersRequestError{
				Request: req,
				Error:   err,
			}
			continue
		}
		containers <- c
	}
	wg.Done()
}

// ParallelContainers creates containers with parameters and run them in parallel mode
func ParallelContainers(ctx context.Context, reqs ParallelRequest, opt ParallelContainersOptions) ([]StartedContainer, error) {
	if opt.WorkersCount == 0 {
		opt.WorkersCount = defaultWorkersCount
	}

	tasksChanSize := opt.WorkersCount
	if tasksChanSize > len(reqs) {
		tasksChanSize = len(reqs)
	}

	tasksChan := make(chan Request, tasksChanSize)
	errsChan := make(chan ParallelContainersRequestError)
	resChan := make(chan StartedContainer)
	waitRes := make(chan struct{})

	containers := make([]StartedContainer, 0)
	errors := make([]ParallelContainersRequestError, 0)

	wg := sync.WaitGroup{}
	wg.Add(tasksChanSize)

	// run workers
	for i := 0; i < tasksChanSize; i++ {
		go parallelContainersRunner(ctx, tasksChan, errsChan, resChan, &wg)
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

	if len(errors) != 0 {
		return containers, ParallelContainersError{Errors: errors}
	}

	return containers, nil
}
