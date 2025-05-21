package kafka

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Kafka container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithClusterID(clusterID string) Option {
	return func(o *options) error {
		o.env["CLUSTER_ID"] = clusterID

		return nil
	}
}

// configureControllerQuorumVoters sets the quorum voters for the controller. For that, it will
// check if there are any network aliases defined for the container and use the first alias in the
// first network. Else, it will use localhost.
func configureControllerQuorumVoters() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}

		if req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] == "" {
			host := "localhost"
			if len(req.Networks) > 0 {
				nw := req.Networks[0]
				if len(req.NetworkAliases[nw]) > 0 {
					host = req.NetworkAliases[nw][0]
				}
			}

			req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"] = fmt.Sprintf("1@%s:9094", host)
		}

		return nil
	}
}
