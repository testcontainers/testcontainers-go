package testcontainers

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	"github.com/testcontainers/testcontainers-go/internal/core/reaper"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

// NewNetwork creates a new network with a random UUID name.
// By default, the network is created with the following options:
// - Driver: bridge
// - Labels: the Testcontainers for Go generic labels, to be managed by Ryuk. Please see the GenericLabels() function
// And those options can be modified by the user, using the CreateModifier function field.
func NewNetwork(ctx context.Context, opts ...tcnetwork.Customizer) (*DockerNetwork, error) {
	nc := types.NetworkCreate{
		Driver: "bridge",
		Labels: core.DefaultLabels(core.SessionID()),
	}

	for _, opt := range opts {
		if err := opt.Customize(&nc); err != nil {
			return nil, err
		}
	}

	// Make sure that bridge network exists
	// In case it is disabled we will create reaper_default network
	if _, err := corenetwork.GetDefault(ctx); err != nil {
		return nil, fmt.Errorf("default network not found: %w", err)
	}

	tcConfig := config.Read()

	var err error
	var termSignal chan bool

	if !tcConfig.RyukDisabled {
		_, err := NewReaper(context.Background(), core.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to create reaper: %w", err)
		}

		termSignal, err = reaper.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect network to reaper: %w", err)
		}
	}

	// Cleanup on error, otherwise set termSignal to nil before successful return.
	defer func() {
		if termSignal != nil {
			termSignal <- true
		}
	}()

	req := corenetwork.Request{
		Driver:         nc.Driver,
		CheckDuplicate: nc.CheckDuplicate, //nolint:staticcheck
		Internal:       nc.Internal,
		EnableIPv6:     nc.EnableIPv6,
		Name:           uuid.NewString(),
		Labels:         nc.Labels,
		Attachable:     nc.Attachable,
		IPAM:           nc.IPAM,
	}

	response, err := corenetwork.New(ctx, req)
	if err != nil {
		return &DockerNetwork{}, err
	}

	n := &DockerNetwork{
		ID:                response.ID,
		Driver:            req.Driver,
		Name:              req.Name,
		terminationSignal: termSignal,
	}

	// Disable cleanup on success
	termSignal = nil

	return n, nil
}
