package testcontainers

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
)

func (c *DockerContainer) GetContainerID() string {
	return c.ID
}

// startLogProduction will start a concurrent process that will continuously read logs
// from the container and will send them to each added LogConsumer.
// Default log production timeout is 5s. It is used to set the context timeout
// which means that each log-reading loop will last at least the specified timeout
// and that it cannot be cancelled earlier.
// Use functional option WithLogProductionTimeout() to override default timeout. If it's
// lower than 5s and greater than 60s it will be set to 5s or 60s respectively.
func (c *DockerContainer) StartLogProduction(ctx context.Context, logConfig log.ConsumerConfig) error {
	c.logProductionStop = make(chan struct{}, 1) // buffered channel to avoid blocking
	c.logProductionWaitGroup.Add(1)

	for _, opt := range logConfig.Opts {
		opt(c)
	}

	minLogProductionTimeout := time.Duration(5 * time.Second)
	maxLogProductionTimeout := time.Duration(60 * time.Second)

	if c.logProductionTimeout == nil {
		c.logProductionTimeout = &minLogProductionTimeout
	}

	if *c.logProductionTimeout < minLogProductionTimeout {
		c.logProductionTimeout = &minLogProductionTimeout
	}

	if *c.logProductionTimeout > maxLogProductionTimeout {
		c.logProductionTimeout = &maxLogProductionTimeout
	}

	c.logProductionError = make(chan error, 1)

	go func() {
		defer func() {
			close(c.logProductionError)
			c.logProductionWaitGroup.Done()
		}()

		since := ""
		// if the socket is closed we will make additional logs request with updated Since timestamp
	BEGIN:
		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Since:      since,
		}

		ctx, cancel := context.WithTimeout(ctx, *c.logProductionTimeout)
		defer cancel()

		cli, err := core.NewClient(ctx)
		if err != nil {
			c.logProductionError <- err
			return
		}
		defer cli.Close()

		r, err := cli.ContainerLogs(ctx, c.GetContainerID(), options)
		if err != nil {
			c.logProductionError <- err
			return
		}

		for {
			select {
			case <-c.logProductionStop:
				c.logProductionError <- r.Close()
				return
			default:
			}
			h := make([]byte, 8)
			_, err := io.ReadFull(r, h)
			if err != nil {
				switch {
				case err == io.EOF:
					// No more logs coming
				case errors.Is(err, net.ErrClosed):
					now := time.Now()
					since = fmt.Sprintf("%d.%09d", now.Unix(), int64(now.Nanosecond()))
					goto BEGIN
				case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
					// Probably safe to continue here
					continue
				default:
					_, _ = fmt.Fprintf(os.Stderr, "container log error: %+v. %s", err, log.StoppedForOutOfSyncMessage)
					// if we would continue here, the next header-read will result into random data...
				}
				return
			}

			count := binary.BigEndian.Uint32(h[4:])
			if count == 0 {
				continue
			}
			logType := h[0]
			if logType > 2 {
				_, _ = fmt.Fprintf(os.Stderr, "received invalid log type: %d", logType)
				// sometimes docker returns logType = 3 which is an undocumented log type, so treat it as stdout
				logType = 1
			}

			// a map of the log type --> int representation in the header, notice the first is blank, this is stdin, but the go docker client doesn't allow following that in logs
			logTypes := []string{"", log.Stdout, log.Stderr}

			b := make([]byte, count)
			_, err = io.ReadFull(r, b)
			if err != nil {
				// TODO: add-logger: use logger to log out this error
				_, _ = fmt.Fprintf(os.Stderr, "error occurred reading log with known length %s", err.Error())
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					// Probably safe to continue here
					continue
				}
				// we can not continue here as the next read most likely will not be the next header
				_, _ = fmt.Fprintln(os.Stderr, log.StoppedForOutOfSyncMessage)
				return
			}

			if logConfig.Consumer != nil {
				logConfig.Consumer.Accept(log.Log{
					LogType: logTypes[logType],
					Content: b,
				})
			}
		}
	}()

	return nil
}

// GetLogProductionErrorChannel exposes the only way for the consumer
// to be able to listen to errors and react to them.
func (c *DockerContainer) GetLogProductionErrorChannel() <-chan error {
	return c.logProductionError
}

// StopLogProduction will stop the concurrent process that is reading logs
// and sending them to each added LogConsumer
func (c *DockerContainer) StopLogProduction() error {
	// signal the log production to stop
	c.logProductionStop <- struct{}{}

	c.logProductionWaitGroup.Wait()

	return <-c.logProductionError
}

func (c *DockerContainer) WithLogProductionTimeout(timeout time.Duration) {
	c.logProductionTimeout = &timeout
}
