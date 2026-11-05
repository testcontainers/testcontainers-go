// Package cockroachdbexamples provides the optional pgx adapter for CockroachDB.
package cockroachdbexamples

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

// Container adds pgx configuration and registered database/sql connection
// handles to a parent container. Provisioning remains in cockroachdb.Run.
type Container struct {
	*cockroachdb.CockroachDBContainer
}

// NewContainer wraps a container with the optional pgx adapter.
func NewContainer(c *cockroachdb.CockroachDBContainer) *Container {
	return &Container{CockroachDBContainer: c}
}

// ConnectionConfig returns the pgx config, including the in-memory CA and
// client certificate and the container's TLS server name (127.0.0.1 by default).
func (c *Container) ConnectionConfig(ctx context.Context) (*pgx.ConnConfig, error) {
	info, err := c.ConnectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := pgx.ParseConfig(info.URL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.TLSConfig = info.TLSConfig
	return cfg, nil
}

// ConnectionString returns a registered pgx stdlib handle for sql.Open("pgx/v5", ...).
// It is not a PostgreSQL URL and cannot be passed to pgx.ParseConfig.
// This package registers the driver. Callers own closing connections and may
// release unused handles with stdlib.UnregisterConnConfig.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	cfg, err := c.ConnectionConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("connection config: %w", err)
	}
	return stdlib.RegisterConnConfig(cfg), nil
}

// MustConnectionString returns a registered pgx stdlib handle, panicking on error.
func (c *Container) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}
