package kafka

import (
	"errors"

	"github.com/testcontainers/testcontainers-go"
)

type runOptions struct {
	image         string
	starterScript string
	flavorWasSet  bool
}

// RunOption is an option that configures how Kafka container is started.
type RunOption func(*runOptions) error

var _ testcontainers.ContainerCustomizer = (RunOption)(nil)

func (o RunOption) Customize(_ *testcontainers.GenericContainerRequest) error {
	return nil
}

// WithStarterScript is an option to set a custom starter script content for the Kafka container.
//
// You would typically use this option when the image you are using is different from
// the standard ones and the default starter script does not work as expected.
// This option conflicts with WithApacheFlavor and WithConfluentFlavor options,
// and the last one provided takes precedence.
func WithStarterScript(content string) RunOption {
	return func(o *runOptions) error {
		o.starterScript = content
		return nil
	}
}

func (o *runOptions) getStarterScriptContent() string {
	if o.starterScript == "" {
		if isApache(o.image) {
			return apacheStarterScript
		}
		// Default to confluentinc for backward compatibility
		// in situations when image was custom specified based on confluentinc
		return confluentStarterScript
	}
	return o.starterScript
}

// WithClusterID sets the CLUSTER_ID environment variable for the Kafka container.
func WithClusterID(clusterID string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"CLUSTER_ID": clusterID,
	})
}

var errFlavorAlreadySet = errors.New("flavor was already set, provide only one of WithApacheFlavor or WithConfluentFlavor")

// WithApacheFlavor sets the starter script to the one compatible with Apache Kafka images.
//
// Note: this option conflicts with WithConfluentFlavor option, and the error is returned
// if both are provided. The option also conflicts with WithStarterScript option,
// but in that case the last one provided takes precedence.
func WithApacheFlavor() RunOption {
	return func(o *runOptions) error {
		if o.flavorWasSet {
			return errFlavorAlreadySet
		}
		o.flavorWasSet = true
		o.starterScript = apacheStarterScript
		return nil
	}
}

// WithConfluentFlavor sets the starter script to the one compatible with Confluent Kafka images.
//
// Note: this option conflicts with WithApacheFlavor option, and the error is returned
// if both are provided. The option also conflicts with WithStarterScript option,
// but in that case the last one provided takes precedence.
func WithConfluentFlavor() RunOption {
	return func(o *runOptions) error {
		if o.flavorWasSet {
			return errFlavorAlreadySet
		}
		o.flavorWasSet = true
		o.starterScript = confluentStarterScript
		return nil
	}
}
