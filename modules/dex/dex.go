// Package dex provides a testcontainers module for the Dex OIDC provider.
//
// Supported grants: authorization_code, refresh_token, password. The
// client_credentials grant requires Dex ≥ v2.46.0 (or dexidp/dex:master)
// together with WithEnableClientCredentials() — this sets the
// DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true env var that gates
// the feature. Earlier releases return unsupported_grant_type.
//
// Example:
//
//	ctx := context.Background()
//	app, err := dex.NewClient("my-app",
//	    dex.WithClientSecret("s3cr3t"),
//	    dex.WithClientRedirectURIs("http://localhost/callback"),
//	)
//	if err != nil { log.Fatal(err) }
//	user, err := dex.NewUser("u@example.com", "u", "p")
//	if err != nil { log.Fatal(err) }
//
//	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
//	    dex.WithClient(app),
//	    dex.WithUser(user),
//	)
//	defer testcontainers.TerminateContainer(c)
package dex

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	httpPort = "5556/tcp"
	grpcPort = "5557/tcp"

	configPath = "/etc/dex/dex.yml"
)

// Container is a running Dex OIDC provider.
type Container struct {
	testcontainers.Container
	issuer string
}

// Run starts Dex. The image is required (tc-go convention). Module options
// (WithClient, WithUser, WithIssuer, ...) and generic tc-go customizers may
// be mixed in the opts slice.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("dex: apply option: %w", err)
			}
		}
	}

	container := &Container{}

	postStart := func(ctx context.Context, c testcontainers.Container) error {
		if settings.issuer == "" {
			host, err := c.Host(ctx)
			if err != nil {
				return fmt.Errorf("dex: host: %w", err)
			}
			mapped, err := c.MappedPort(ctx, httpPort)
			if err != nil {
				return fmt.Errorf("dex: mapped port: %w", err)
			}
			settings.issuer = fmt.Sprintf("http://%s:%s/dex", host, mapped.Port())
		}
		container.issuer = settings.issuer

		yml, err := render(settings)
		if err != nil {
			return fmt.Errorf("dex: render yaml: %w", err)
		}
		if err := c.CopyToContainer(ctx, yml, configPath, 0o644); err != nil {
			return fmt.Errorf("dex: copy yaml: %w", err)
		}

		// Wait for Dex to serve discovery + gRPC port before returning. The
		// discovery path is derived from the issuer's path component so
		// WithIssuer("http://host/idp") probes "/idp/.well-known/...", not
		// the default "/dex/...".
		issuerURL, err := url.Parse(settings.issuer)
		if err != nil {
			return fmt.Errorf("dex: parse issuer: %w", err)
		}
		discoveryPath := strings.TrimRight(issuerURL.Path, "/") + "/.well-known/openid-configuration"
		ready := wait.ForAll(
			wait.ForHTTP(discoveryPath).
				WithPort(httpPort).
				WithStartupTimeout(60*time.Second),
			wait.ForListeningPort(grpcPort).
				WithStartupTimeout(60*time.Second),
		)
		return ready.WaitUntilReady(ctx, c)
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(httpPort, grpcPort),
		testcontainers.WithEntrypoint("/bin/sh"),
		testcontainers.WithCmd("-c",
			"while [ ! -f "+configPath+" ]; do sleep 0.1; done; "+
				"exec dex serve "+configPath),
		testcontainers.WithLifecycleHooks(testcontainers.ContainerLifecycleHooks{
			PostStarts: []testcontainers.ContainerHook{postStart},
		}),
		// No wait strategy on the request — Dex would fail readiness before the
		// PostStart hook copies the YAML. The hook performs its own wait after
		// copying.
	}
	if settings.logger != nil {
		moduleOpts = append(moduleOpts,
			testcontainers.WithLogConsumers(newSlogConsumer(settings.logger)))
	}
	if settings.enableClientCredentials {
		moduleOpts = append(moduleOpts,
			testcontainers.WithEnv(map[string]string{
				"DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT": "true",
			}))
	}
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		container.Container = ctr
	}
	if err != nil {
		return container, fmt.Errorf("dex: run: %w", err)
	}
	return container, nil
}

// IssuerURL returns Dex's issuer URL. Empty if Run has not started.
func (c *Container) IssuerURL() string { return c.issuer }

// ConfigEndpoint returns the OIDC discovery document URL.
func (c *Container) ConfigEndpoint() string {
	return c.issuer + "/.well-known/openid-configuration"
}

// JWKSEndpoint returns the JSON Web Key Set URL.
func (c *Container) JWKSEndpoint() string { return c.issuer + "/keys" }

// TokenEndpoint returns the OAuth2 token URL.
func (c *Container) TokenEndpoint() string { return c.issuer + "/token" }

// AuthEndpoint returns the OAuth2 authorization URL.
func (c *Container) AuthEndpoint() string { return c.issuer + "/auth" }

// GRPCEndpoint returns host:mappedPort for Dex's gRPC admin API. Errors
// propagate from the Docker API, or report that the container has not been
// started.
func (c *Container) GRPCEndpoint(ctx context.Context) (string, error) {
	if c.Container == nil {
		return "", errors.New("dex: container not started")
	}
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("dex: host: %w", err)
	}
	port, err := c.MappedPort(ctx, grpcPort)
	if err != nil {
		return "", fmt.Errorf("dex: mapped grpc port: %w", err)
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
