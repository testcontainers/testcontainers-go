package testcontainers

import "github.com/testcontainers/testcontainers-go/internal/logging"

// SetLogger sets the default logger
//
// Deprecated: testcontainers is changing how logging works, this function is a workaround
// whilst the future functionality is implemented
func SetLogger(l logging.Logging) {
	logging.Logger = l
}
