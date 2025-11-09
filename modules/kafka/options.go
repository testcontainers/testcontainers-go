package kafka

import "github.com/testcontainers/testcontainers-go"

type runOptions struct {
	image         string
	starterScript string
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
// the standard ones or the image is in your custom registry and automatic inference
// of the starter script does not work as expected.
func WithStarterScript(content string) RunOption {
	return func(o *runOptions) error {
		o.starterScript = content
		return nil
	}
}

func (o *runOptions) getStarterScriptContent() string {
	if o.starterScript == "" {
		if isApache(o.image) {
			return ApacheStarterScript
		}
		// Default to confluentinc for backward compatibility
		// in situations when image was custom specified based on confluentinc
		return ConfluentStarterScript
	}
	return o.starterScript
}
