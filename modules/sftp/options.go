package sftp

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// options holds the configuration for the SFTP container.
type options struct {
	// users holds the list of user configuration strings passed as CMD args.
	// Each entry is in the form "username:password:::" (empty dir field;
	// atmoz/sftp defaults to /home/<username>/upload).
	users []string
}

func defaultOptions() *options {
	return &options{}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the SFTP container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithUser adds an SFTP user with the given username and password.
// The user's home directory will default to /home/<username>/upload.
// Multiple calls to WithUser will accumulate users.
func WithUser(username, password string) Option {
	return func(o *options) {
		o.users = append(o.users, fmt.Sprintf("%s:%s:::", username, password))
	}
}
