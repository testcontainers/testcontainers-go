package kafka

import "github.com/testcontainers/testcontainers-go"

type runOptions struct {
	image         string
	starterScript string
}

type Option func(*runOptions) error

var _ testcontainers.ContainerCustomizer = (Option)(nil)

func (o Option) Customize(req *testcontainers.GenericContainerRequest) error {
	return nil
}

func WithStarterScript(content string) Option {
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
