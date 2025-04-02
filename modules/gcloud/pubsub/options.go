package pubsub

import "github.com/testcontainers/testcontainers-go/modules/gcloud/internal/shared"

// Options aliases the common GCloud options
type options = shared.Options

// Option aliases the common GCloud option type
type Option = shared.Option

// defaultOptions returns a new Options instance with the default project ID.
func defaultOptions() options {
	return shared.DefaultOptions()
}

// WithProjectID re-exports the common GCloud WithProjectID option
var WithProjectID = shared.WithProjectID
