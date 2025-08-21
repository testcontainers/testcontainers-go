package nebulagraph

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// NebulaGraphContainer represents a running NebulaGraph cluster for testing
type NebulaGraphContainer struct {
	Graphd   testcontainers.Container
	Metad    testcontainers.Container
	Storaged testcontainers.Container
}

func RunContainer(ctx context.Context) (*NebulaGraphContainer, error) {
	graphdCustomizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage(defaultGraphdImage),
	}

	metadCustomizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage(defaultMetadImage),
	}

	storagedCustomizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage(defaultStoragedImage),
	}
	return Run(ctx, graphdCustomizers, storagedCustomizers, metadCustomizers)
}

// Run starts NebulaGraph (metad, storaged, graphd) containers for testing
func Run(ctx context.Context,
	graphdCustomizers []testcontainers.ContainerCustomizer,
	storagedCustomizers []testcontainers.ContainerCustomizer,
	metadCustomizers []testcontainers.ContainerCustomizer,
) (*NebulaGraphContainer, error) {
	// 1. Create a custom network
	netRes, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	// 2. Start metad
	metadReq := defaultMetadContainerRequest(netRes.Name)
	for _, customizer := range metadCustomizers {
		if err := customizer.Customize(&metadReq); err != nil {
			return nil, err
		}
	}

	metad, err := testcontainers.GenericContainer(ctx, metadReq)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start metad: %w", err)}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}
		return nil, errors.Join(errs...)
	}

	// Wait for metad to be healthy before starting storaged
	if err := waitForHTTPReady(ctx, metad, metadPortHTTP); err != nil {
		errs := []error{fmt.Errorf("metad not healthy: %w", err)}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}
		return nil, errors.Join(errs...)
	}

	// 3. Start graphd (needed for storage registration)
	graphdReq := defaultGrapdContainerRequest(netRes.Name)
	for _, customizer := range graphdCustomizers {
		if err := customizer.Customize(&graphdReq); err != nil {
			return nil, err
		}
	}

	graphd, err := testcontainers.GenericContainer(ctx, graphdReq)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start graphd: %w", err)}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}
		return nil, errors.Join(errs...)
	}

	// Wait for graphd to be healthy before starting storaged
	if err := waitForHTTPReady(ctx, graphd, graphdPortHTTP); err != nil {
		errs := []error{fmt.Errorf("graphd not healthy: %w", err)}

		if err := graphd.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("graphd terminate: %w", err))
		}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}
		return nil, errors.Join(errs...)
	}

	// 4. Start storaged
	storagedReq := defaultStoragedContainerRequest(netRes.Name)
	for _, customizer := range storagedCustomizers {
		if err := customizer.Customize(&storagedReq); err != nil {
			return nil, err
		}
	}
	storaged, err := testcontainers.GenericContainer(ctx, storagedReq)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start storaged: %w", err)}

		if err := graphd.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("graphd terminate: %w", err))
		}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}

		return nil, errors.Join(errs...)
	}
	// 5. Run storage registration command with retry logic
	activator, err := testcontainers.GenericContainer(ctx, defaultActivatorContainerRequest(netRes.Name))
	if err != nil {
		errs := []error{fmt.Errorf("failed to start activator container: %w", err)}

		if err := storaged.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("storaged terminate: %w", err))
		}
		if err := graphd.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("graphd terminate: %w", err))
		}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}

		return nil, errors.Join(errs...)
	}

	// Create a channel to signal when activation is successful
	activationDone := make(chan struct{})

	go func() {
		_ = activator.Start(ctx)
	}()

	// Start a goroutine to watch the activator logs
	go func() {
		defer close(activationDone)
		for {
			logs, err := activator.Logs(ctx)
			if err != nil {
				return
			}
			defer logs.Close()

			buf := make([]byte, 1024)
			for {
				n, err := logs.Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					output := string(buf[:n])
					if strings.Contains(output, "✔️ Storage activated successfully") ||
						strings.Contains(output, "✔️ Storage already activated") {
						_ = activator.Terminate(ctx)
						return
					}
				}
			}
		}
	}()

	// Wait for either activation success or timeout
	select {
	case <-activationDone:
		// Success case - continue with the next steps
	case <-time.After(3 * time.Minute):
		errs := []error{errors.New("storage registration timed out")}
		if err := activator.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("activator terminate: %w", err))
		}
		if err := storaged.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("storaged terminate: %w", err))
		}
		if err := graphd.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("graphd terminate: %w", err))
		}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}

		errs = append(errs, errors.New("storage registration timed out"))

		return nil, errors.Join(errs...)
	}

	// Give meta service a moment to process the registration
	time.Sleep(2 * time.Second)

	// 6. Wait for storaged to be healthy
	time.Sleep(2 * time.Second) // Small pause after successful registration

	if err = waitForHTTPReady(ctx, storaged, storagedPortHTTP); err != nil {
		errs := []error{fmt.Errorf("storaged not healthy: %w", err)}

		if err := storaged.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("storaged terminate: %w", err))
		}
		if err := graphd.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("graphd terminate: %w", err))
		}
		if err := metad.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metad terminate: %w", err))
		}
		if err := netRes.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}

		errs = append(errs, errors.New("storage registration timed out"))

		return nil, errors.Join(errs...)
	}

	return &NebulaGraphContainer{
		Graphd:   graphd,
		Metad:    metad,
		Storaged: storaged,
	}, nil
}

// waitForHTTPReady waits for the given container's HTTP status endpoint to be ready on the mapped port
func waitForHTTPReady(ctx context.Context, container testcontainers.Container, containerPort string) error {
	host, err := container.Host(ctx)
	if err != nil {
		return err
	}
	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%s/status", host, mappedPort.Port())
	timeout := time.Now().Add(2 * time.Minute)
	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for %s", url)
		}
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ConnectionString returns the host:port for connecting to NebulaGraph graphd
func (c *NebulaGraphContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Graphd.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.Graphd.MappedPort(ctx, "9669")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// Terminate stops all NebulaGraph containers
func (c *NebulaGraphContainer) Terminate(ctx context.Context) error {
	var err1, err2, err3 error
	err1 = c.Graphd.Terminate(ctx)
	err2 = c.Storaged.Terminate(ctx)
	err3 = c.Metad.Terminate(ctx)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	return nil
}
