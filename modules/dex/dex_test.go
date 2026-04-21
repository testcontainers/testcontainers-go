package dex_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	dexImage       = "dexidp/dex:v2.45.1"
	dexImageWithCC = "dexidp/dex:master"
)

func TestRun_DefaultPath_DiscoveryMatches(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithUser(dex.User{Email: "u@e.com", Username: "u", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	assert.NotEmpty(t, c.IssuerURL())
	assert.Equal(t, c.IssuerURL()+"/.well-known/openid-configuration", c.ConfigEndpoint())
	assert.Equal(t, c.IssuerURL()+"/keys", c.JWKSEndpoint())
	assert.Equal(t, c.IssuerURL()+"/token", c.TokenEndpoint())
	assert.Equal(t, c.IssuerURL()+"/auth", c.AuthEndpoint())
	assert.NotEmpty(t, c.GRPCEndpoint())

	resp, err := http.Get(c.ConfigEndpoint())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode, "discovery endpoint must return 200")

	var doc map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))

	assert.Equal(t, c.IssuerURL(), doc["issuer"])
	assert.Equal(t, c.JWKSEndpoint(), doc["jwks_uri"])
	assert.Equal(t, c.TokenEndpoint(), doc["token_endpoint"])
	assert.Equal(t, c.AuthEndpoint(), doc["authorization_endpoint"])
}

func TestRun_WithIssuerOverride(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const issuer = "http://dex-override.test:5556/dex"

	c, err := dex.Run(ctx, dexImage,
		dex.WithIssuer(issuer),
		dex.WithUser(dex.User{Email: "u@e.com", Username: "u", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	assert.Equal(t, issuer, c.IssuerURL())
	assert.Equal(t, issuer+"/.well-known/openid-configuration", c.ConfigEndpoint())

	// Cross-check: discovery doc MUST echo the overridden issuer, proving
	// the YAML was rendered with the override and Dex booted against it.
	host, err := c.Host(ctx)
	require.NoError(t, err)
	mapped, err := c.MappedPort(ctx, "5556/tcp")
	require.NoError(t, err)

	reachable := fmt.Sprintf("http://%s:%s/dex/.well-known/openid-configuration", host, mapped.Port())
	resp, err := http.Get(reachable)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var doc map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	assert.Equal(t, issuer, doc["issuer"], "discovery doc must echo the overridden issuer")
}

func TestGRPC_AddRemoveClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithUser(dex.User{Email: "u@e.com", Username: "u", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cl := dex.Client{
		ID:           "runtime-app",
		Secret:       "s",
		RedirectURIs: []string{"http://localhost/cb"},
		Name:         "Runtime App",
	}
	require.NoError(t, c.AddClient(ctx, cl))

	// Idempotency: second Add returns ErrClientExists.
	err = c.AddClient(ctx, cl)
	assert.ErrorIs(t, err, dex.ErrClientExists)

	// Removal succeeds.
	require.NoError(t, c.RemoveClient(ctx, cl.ID))

	// Second remove errors (not-found).
	err = c.RemoveClient(ctx, cl.ID)
	assert.Error(t, err)
}

func TestGRPC_AddRemoveUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	u := dex.User{Email: "runtime@example.com", Username: "runtime", Password: "p"}
	require.NoError(t, c.AddUser(ctx, u))

	// Duplicate add errors.
	err = c.AddUser(ctx, u)
	assert.ErrorIs(t, err, dex.ErrUserExists)

	// Removal succeeds.
	require.NoError(t, c.RemoveUser(ctx, u.Email))

	// Second removal errors.
	err = c.RemoveUser(ctx, u.Email)
	assert.Error(t, err)
}

func TestWithLogger_CapturesDexOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var buf safeBuffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, err := dex.Run(ctx, dexImage,
		dex.WithLogger(logger),
		dex.WithUser(dex.User{Email: "u@e.com", Username: "u", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	// Poll for the "listening on" line that Dex always emits at the end of
	// its startup sequence. This proves the log pipe is fully wired — not
	// just that boot started, but that log output flowed through after the
	// container was declared ready.
	//
	// NOTE: Dex v2.x does not log per-operation gRPC events (CreateClient,
	// CreatePassword, etc.) — only boot and key-rotation events reach the
	// log stream. "listening on" is the last boot-time line and therefore
	// the strongest stable signal we can assert on.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), "listening on") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Fallback: even if the specific string drifted across Dex
	// versions, we MUST have captured at least one line. An empty buffer
	// means the log pipe never wired up.
	require.NotEmpty(t, buf.String(), "logger captured no output — pipe not wired")
	t.Logf("captured logs (first 500 bytes):\n%.500s", buf.String())
	t.Fatalf("expected \"listening on\" in captured output")
}

// safeBuffer is a bytes.Buffer wrapper safe for concurrent writes from
// the LogConsumer goroutine and reads from the test goroutine.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestAuthCode_PasswordConnector_Basic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const redirectURI = "http://localhost:18080/cb"

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID:           "e2e-app",
			Secret:       "e2e-secret",
			RedirectURIs: []string{redirectURI},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
			Name:         "E2E App",
		}),
		dex.WithUser(dex.User{
			Email:    "alice@example.com",
			Username: "alice",
			Password: "pass",
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := oauth2.Config{
		ClientID:     "e2e-app",
		ClientSecret: "e2e-secret",
		RedirectURL:  redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.AuthEndpoint(),
			TokenURL: c.TokenEndpoint(),
		},
		Scopes: []string{"openid", "email", "profile"},
	}

	tok := drivePasswordAuthCode(t, ctx, cfg, "alice@example.com", "pass")
	assert.NotEmpty(t, tok.AccessToken)

	idToken, ok := tok.Extra("id_token").(string)
	require.True(t, ok, "id_token missing from response")
	assert.NotEmpty(t, idToken)
}

func TestAuthCode_RefreshToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const redirectURI = "http://localhost:18080/cb"

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "e2e", Secret: "s", Name: "E2E",
			RedirectURIs: []string{redirectURI},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		}),
		dex.WithUser(dex.User{Email: "a@e.com", Username: "a", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := oauth2.Config{
		ClientID:     "e2e",
		ClientSecret: "s",
		RedirectURL:  redirectURI,
		Endpoint:     oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:       []string{"openid", "offline_access"},
	}

	tok := drivePasswordAuthCode(t, ctx, cfg, "a@e.com", "p")
	require.NotEmpty(t, tok.RefreshToken, "offline_access scope should yield refresh_token")

	// Swap access_token to force a refresh via the token source.
	expired := *tok
	expired.AccessToken = "invalid"
	expired.Expiry = time.Now().Add(-1 * time.Hour)

	src := cfg.TokenSource(ctx, &expired)
	newTok, err := src.Token()
	require.NoError(t, err, "refresh exchange failed")
	assert.NotEqual(t, tok.AccessToken, newTok.AccessToken, "refresh yields new access token")
}

func TestAuthCode_MultipleRedirectURIs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	uris := []string{"http://localhost:18080/cb", "http://localhost:18090/cb"}

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "e2e", Secret: "s", Name: "E2E",
			RedirectURIs: uris,
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		}),
		dex.WithUser(dex.User{Email: "a@e.com", Username: "a", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	for _, uri := range uris {
		cfg := oauth2.Config{
			ClientID:     "e2e",
			ClientSecret: "s",
			RedirectURL:  uri,
			Endpoint:     oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
			Scopes:       []string{"openid"},
		}
		tok := drivePasswordAuthCode(t, ctx, cfg, "a@e.com", "p")
		assert.NotEmpty(t, tok.AccessToken, "uri=%s", uri)
	}
}

func TestClientCredentials_UnsupportedByLocalConnectors(t *testing.T) {
	// Locks in a known Dex limitation: the OAuth2 client_credentials grant
	// is NOT implemented for the built-in password connector or mockCallback
	// in Dex v2.x. Only upstream OIDC/LDAP connectors implement the
	// ConnectorWithClientCredentials interface. This test asserts the
	// documented failure mode so regressions are surfaced early if Dex
	// ever adds local-connector CC support.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "svc", Secret: "s", Name: "Service",
			GrantTypes: []string{"client_credentials"},
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := clientcredentials.Config{
		ClientID:     "svc",
		ClientSecret: "s",
		TokenURL:     c.TokenEndpoint(),
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	_, err = cfg.TokenSource(ctx).Token()
	require.Error(t, err, "Dex should reject CC against local connectors")

	// The error surfaces as an OAuth2 token response error; the wire
	// message is "unsupported_grant_type" or similar. Match loosely so
	// minor Dex wording changes don't flake.
	msg := strings.ToLower(err.Error())
	assert.True(t,
		strings.Contains(msg, "unsupported") || strings.Contains(msg, "grant"),
		"expected grant-related error, got: %v", err)
}

func TestPasswordGrant_ROPC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "cli", Secret: "s", Name: "CLI",
			GrantTypes: []string{"password"},
		}),
		dex.WithUser(dex.User{Email: "a@e.com", Username: "a", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := oauth2.Config{
		ClientID:     "cli",
		ClientSecret: "s",
		Endpoint:     oauth2.Endpoint{TokenURL: c.TokenEndpoint()},
		Scopes:       []string{"openid"},
	}
	tok, err := cfg.PasswordCredentialsToken(ctx, "a@e.com", "p")
	require.NoError(t, err)
	assert.NotEmpty(t, tok.AccessToken)
}

func TestMultipleClients_OneInstance(t *testing.T) {
	// Registers two clients on a single Dex instance and verifies both can
	// independently obtain tokens. The "svc" client uses the password grant
	// (Dex's M2M pattern — see TestClientCredentials); the "web" client uses
	// the authorization_code flow. Tokens from distinct clients must differ.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "svc", Secret: "s", Name: "SVC",
			GrantTypes: []string{"password"},
		}),
		dex.WithClient(dex.Client{
			ID: "web", Secret: "s", Name: "Web",
			RedirectURIs: []string{"http://localhost/cb"},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		}),
		dex.WithUser(dex.User{Email: "a@e.com", Username: "a", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	// Service-account client via password grant.
	svcCfg := oauth2.Config{
		ClientID:     "svc",
		ClientSecret: "s",
		Endpoint:     oauth2.Endpoint{TokenURL: c.TokenEndpoint()},
		Scopes:       []string{"openid"},
	}
	svcTok, err := svcCfg.PasswordCredentialsToken(ctx, "a@e.com", "p")
	require.NoError(t, err)
	require.NotEmpty(t, svcTok.AccessToken)

	// Web client via auth_code flow.
	webCfg := oauth2.Config{
		ClientID: "web", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}
	webTok := drivePasswordAuthCode(t, ctx, webCfg, "a@e.com", "p")
	assert.NotEmpty(t, webTok.AccessToken)

	assert.NotEqual(t, svcTok.AccessToken, webTok.AccessToken, "tokens from distinct clients must differ")
}

func TestMockConnector_IssuesToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithConnector(dex.ConnectorMock, "mock", "Mock Connector"),
		dex.WithClient(dex.Client{
			ID: "e2e", Secret: "s", Name: "E2E",
			RedirectURIs: []string{"http://localhost/cb"},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := oauth2.Config{
		ClientID: "e2e", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}

	// Drive the /auth URL with connector_id=mock so Dex skips the login form.
	authURL := cfg.AuthCodeURL("state-mock") + "&connector_id=mock"
	req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.HasPrefix(req.URL.String(), cfg.RedirectURL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	loc := resp.Request.URL
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		parsed, perr := url.Parse(resp.Header.Get("Location"))
		require.NoError(t, perr)
		loc = parsed
	}
	code := loc.Query().Get("code")
	require.NotEmpty(t, code, "mockCallback should redirect with ?code=...; got %q", loc.String())

	tok, err := cfg.Exchange(ctx, code)
	require.NoError(t, err)
	assert.NotEmpty(t, tok.AccessToken)
}

func TestGRPC_RuntimeAddUsableEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	// Seed client + user at runtime.
	require.NoError(t, c.AddClient(ctx, dex.Client{
		ID: "late-app", Secret: "s",
		RedirectURIs: []string{"http://localhost/cb"},
		Name:         "Late App",
	}))
	require.NoError(t, c.AddUser(ctx, dex.User{
		Email: "late@e.com", Username: "late", Password: "p",
	}))

	cfg := oauth2.Config{
		ClientID: "late-app", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}
	tok := drivePasswordAuthCode(t, ctx, cfg, "late@e.com", "p")
	assert.NotEmpty(t, tok.AccessToken)

	// Remove user — subsequent login attempt must fail.
	require.NoError(t, c.RemoveUser(ctx, "late@e.com"))

	// Do the login dance manually — drivePasswordAuthCode uses require.NoError
	// which would abort the test on the expected failure.
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.HasPrefix(req.URL.String(), cfg.RedirectURL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	authURL := cfg.AuthCodeURL("s1")
	resp, err := client.Get(authURL)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	loginURL := resp.Request.URL.String()

	// Use the same action-extractor approach as the helper — but for a
	// negative path we just POST blindly to the request URL. Dex's
	// local login endpoint accepts POSTs at the same URL the GET returned.
	_ = body

	form := url.Values{"login": {"late@e.com"}, "password": {"p"}}
	r2, err := client.Post(loginURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	body2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()
	// Dex renders the login page again with a failure marker. The exact
	// copy ("Invalid Email Address and password") may drift across
	// versions; match loosely.
	lower := strings.ToLower(string(body2))
	assert.True(t,
		strings.Contains(lower, "invalid") || strings.Contains(lower, "authentication failed") || r2.StatusCode >= 400,
		"removed user should fail login; got status=%d body-prefix=%.200q", r2.StatusCode, lower)
}

func TestWithIssuer_CrossContainerViaNetworkAlias(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	net, err := network.New(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = net.Remove(ctx) })

	const issuer = "http://dex:5556/dex"

	c, err := dex.Run(ctx, dexImage,
		dex.WithIssuer(issuer),
		dex.WithUser(dex.User{Email: "u@e.com", Username: "u", Password: "p"}),
		network.WithNetwork([]string{"dex"}, net),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	// Sidecar: curl the discovery endpoint through the network alias.
	sidecarReq := testcontainers.ContainerRequest{
		Image:          "curlimages/curl:8.10.1",
		Networks:       []string{net.Name},
		NetworkAliases: map[string][]string{net.Name: {"sidecar"}},
		Cmd: []string{
			"sh", "-c",
			"curl -fsS http://dex:5556/dex/.well-known/openid-configuration",
		},
		WaitingFor: wait.ForExit(),
	}
	sidecar, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: sidecarReq,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(sidecar) })

	logs, err := sidecar.Logs(ctx)
	require.NoError(t, err)
	body, err := io.ReadAll(logs)
	require.NoError(t, err)

	assert.Contains(t, string(body), issuer,
		"discovery doc fetched via network alias must echo overridden issuer")
}

func TestClientCredentials_WithFeatureFlag(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImageWithCC,
		dex.WithEnableClientCredentials(),
		dex.WithClient(dex.Client{
			ID: "svc", Secret: "s", Name: "Service",
			GrantTypes: []string{"client_credentials"},
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := clientcredentials.Config{
		ClientID:     "svc",
		ClientSecret: "s",
		TokenURL:     c.TokenEndpoint(),
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	tok, err := cfg.TokenSource(ctx).Token()
	require.NoError(t, err, "CC grant should succeed with feature flag enabled on master image")
	assert.NotEmpty(t, tok.AccessToken)
}

func TestConsumer_IDTokenVerifies_CoreosOIDC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const redirectURI = "http://localhost:18080/cb"

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(dex.Client{
			ID: "e2e", Secret: "s", Name: "E2E",
			RedirectURIs: []string{redirectURI},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
		}),
		dex.WithUser(dex.User{Email: "a@e.com", Username: "a", Password: "p"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })

	cfg := oauth2.Config{
		ClientID: "e2e", ClientSecret: "s",
		RedirectURL: redirectURI,
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid", "email", "profile"},
	}
	tok := drivePasswordAuthCode(t, ctx, cfg, "a@e.com", "p")

	rawIDToken, ok := tok.Extra("id_token").(string)
	require.True(t, ok, "id_token missing from response")

	provider, err := oidc.NewProvider(ctx, c.IssuerURL())
	require.NoError(t, err)

	verifier := provider.Verifier(&oidc.Config{ClientID: "e2e"})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	require.NoError(t, err, "coreos/go-oidc failed to verify Dex-issued token")

	// Dex maps the password-connector username to the "name" claim (profile scope).
	// "preferred_username" is not set in the ID token for the built-in password connector.
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	require.NoError(t, idToken.Claims(&claims))
	assert.Equal(t, "a@e.com", claims.Email)
	assert.Equal(t, "a", claims.Name)
	assert.NotEmpty(t, claims.Sub)
}
