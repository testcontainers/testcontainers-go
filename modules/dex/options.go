package dex

import "github.com/testcontainers/testcontainers-go"

// WithIssuer sets the issuer for the Dex container.
// In most cases, the issuer should be set to the URL of the Dex container.
// If not provided, the default issuer will be used, that will not necessarily match the actual URL of the Dex container.
func WithIssuer(issuer string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DEX_ISSUER"] = issuer
		return nil
	})
}

// WithLogLevel overrides the default log level for the Dex container.
// See [log/slog](log/slog#Level) for possible values.
func WithLogLevel(level string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DEX_LOG_LEVEL"] = level
		return nil
	})
}
