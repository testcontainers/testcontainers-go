// Package dex provides a testcontainers module for the Dex OIDC provider.
//
// Supported grants: authorization_code, refresh_token, password. The
// client_credentials grant requires Dex ≥ v2.46.0 (or dexidp/dex:master)
// with WithEnableClientCredentials() — this sets the
// DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true env var that gates the
// feature. Earlier releases (v2.45.x and below) return unsupported_grant_type.
//
// Example:
//
//	ctx := context.Background()
//	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
//	    dex.WithClient(dex.Client{ID: "my-app", Secret: "s3cr3t", RedirectURIs: []string{"http://localhost/callback"}}),
//	    dex.WithUser(dex.User{Email: "u@example.com", Username: "u", Password: "p"}),
//	)
//	defer testcontainers.TerminateContainer(c)
package dex

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

// DexContainer is a running Dex OIDC provider.
type DexContainer struct {
	testcontainers.Container
	issuer string
}

// Run starts Dex. The image is required (tc-go convention). Module options
// (WithClient, WithUser, WithIssuer, ...) and generic tc-go customizers may
// be mixed in the opts slice.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DexContainer, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
	}

	container := &DexContainer{}

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

		// Wait for Dex to serve discovery + gRPC port before returning.
		ready := wait.ForAll(
			wait.ForHTTP("/dex/.well-known/openid-configuration").
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
		if tag := parseImageTag(img); tag != "" && !imageSupportsClientCredentials(tag) {
			warnLogger := settings.logger
			if warnLogger == nil {
				warnLogger = slog.Default()
			}
			warnLogger.Warn("dex: client_credentials grant requested but image tag predates feature flag support; token exchanges will fail",
				slog.String("image", img),
				slog.String("tag", tag),
				slog.String("minimum_tag", "v2.46.0"),
				slog.String("workaround", "use dexidp/dex:master or dexidp/dex:latest"))
		}
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
func (c *DexContainer) IssuerURL() string { return c.issuer }

// ConfigEndpoint returns the OIDC discovery document URL.
func (c *DexContainer) ConfigEndpoint() string {
	return c.issuer + "/.well-known/openid-configuration"
}

// JWKSEndpoint returns the JSON Web Key Set URL.
func (c *DexContainer) JWKSEndpoint() string { return c.issuer + "/keys" }

// TokenEndpoint returns the OAuth2 token URL.
func (c *DexContainer) TokenEndpoint() string { return c.issuer + "/token" }

// AuthEndpoint returns the OAuth2 authorization URL.
func (c *DexContainer) AuthEndpoint() string { return c.issuer + "/auth" }

// GRPCEndpoint returns host:mappedPort for Dex's gRPC admin API. Empty
// before Run has started.
func (c *DexContainer) GRPCEndpoint() string {
	if c.Container == nil {
		return ""
	}
	host, err := c.Host(context.Background())
	if err != nil {
		return ""
	}
	port, err := c.MappedPort(context.Background(), grpcPort)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", host, port.Port())
}

// grpcEndpoint resolves host:mappedPort for Dex's gRPC admin API. Unlike
// GRPCEndpoint(), it propagates Docker API errors so callers can tell a
// pre-start container apart from a Docker-layer failure mid-test.
func (c *DexContainer) grpcEndpoint(ctx context.Context) (string, error) {
	if c.Container == nil {
		return "", fmt.Errorf("dex: container not started")
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

// parseImageTag extracts the tag portion from a Docker image reference like
// "dexidp/dex:v2.45.1" or "ghcr.io/dex:v2.45.1-alpine". Returns empty
// string when no tag is present or the image uses a digest reference.
func parseImageTag(img string) string {
	// Drop any digest suffix; tag lives after the LAST colon that comes
	// after the LAST slash (so we don't mistake the port in a registry
	// like localhost:5000/dex for a tag).
	if idx := strings.LastIndex(img, "@"); idx >= 0 {
		img = img[:idx]
	}
	lastSlash := strings.LastIndex(img, "/")
	colon := strings.LastIndex(img[lastSlash+1:], ":")
	if colon < 0 {
		return ""
	}
	return img[lastSlash+1+colon+1:]
}

// imageSupportsClientCredentials reports whether a Dex image tag is known
// to include the client_credentials feature flag.
//
// Known-good tags:
//   - master, latest (always rebuild from HEAD)
//   - any tag ≥ v2.46.0 once released
//
// Known-bad tags: v2.45.x and earlier.
// Unknown tags (custom builds, non-semver) return true — the caller is
// presumed to know what they're doing.
func imageSupportsClientCredentials(tag string) bool {
	// Strip common tag suffixes the Dex image uses: -alpine, -distroless.
	base := tag
	for _, suffix := range []string{"-alpine", "-distroless"} {
		base = strings.TrimSuffix(base, suffix)
	}
	switch base {
	case "master", "latest":
		return true
	}
	// Parse semver vX.Y.Z. Accept anything ≥ v2.46.0.
	if !strings.HasPrefix(base, "v") {
		// Non-semver custom build — give the caller the benefit of the doubt.
		return true
	}
	parts := strings.SplitN(strings.TrimPrefix(base, "v"), ".", 3)
	if len(parts) < 2 {
		return true // odd shape — trust the caller
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return true
	}
	switch {
	case major > 2:
		return true
	case major < 2:
		return false
	default: // major == 2
		return minor >= 46
	}
}
