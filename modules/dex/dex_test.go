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
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dexImage       = "dexidp/dex:v2.45.1"
	dexImageWithCC = "dexidp/dex:master"
)

// mustClient + mustUser are test helpers: the module's NewClient / NewUser
// constructors return (Client, error), but every test here uses valid input.
// Wrapping them with require lets the test read like the old field-literal
// form while preserving constructor validation.
func mustClient(t *testing.T, id string, opts ...dex.ClientOption) dex.Client {
	t.Helper()
	c, err := dex.NewClient(id, opts...)
	require.NoError(t, err)
	return c
}

func mustUser(t *testing.T, email, username, password string, opts ...dex.UserOption) dex.User {
	t.Helper()
	u, err := dex.NewUser(email, username, password, opts...)
	require.NoError(t, err)
	return u
}

func TestRun_DefaultPath_DiscoveryMatches(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithUser(mustUser(t, "u@e.com", "u", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	require.NotEmpty(t, c.IssuerURL())
	require.Equal(t, c.IssuerURL()+"/.well-known/openid-configuration", c.ConfigEndpoint())
	require.Equal(t, c.IssuerURL()+"/keys", c.JWKSEndpoint())
	require.Equal(t, c.IssuerURL()+"/token", c.TokenEndpoint())
	require.Equal(t, c.IssuerURL()+"/auth", c.AuthEndpoint())
	grpcEP, err := c.GRPCEndpoint(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, grpcEP)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.ConfigEndpoint(), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode, "discovery endpoint must return 200")

	var doc map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))

	require.Equal(t, c.IssuerURL(), doc["issuer"])
	require.Equal(t, c.JWKSEndpoint(), doc["jwks_uri"])
	require.Equal(t, c.TokenEndpoint(), doc["token_endpoint"])
	require.Equal(t, c.AuthEndpoint(), doc["authorization_endpoint"])
}

func TestRun_WithIssuerOverride(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const issuer = "http://dex-override.test:5556/dex"

	c, err := dex.Run(ctx, dexImage,
		dex.WithIssuer(issuer),
		dex.WithUser(mustUser(t, "u@e.com", "u", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	require.Equal(t, issuer, c.IssuerURL())
	require.Equal(t, issuer+"/.well-known/openid-configuration", c.ConfigEndpoint())

	// Cross-check: discovery doc MUST echo the overridden issuer, proving
	// the YAML was rendered with the override and Dex booted against it.
	host, err := c.Host(ctx)
	require.NoError(t, err)
	mapped, err := c.MappedPort(ctx, "5556/tcp")
	require.NoError(t, err)

	reachable := fmt.Sprintf("http://%s:%s/dex/.well-known/openid-configuration", host, mapped.Port())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reachable, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var doc map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, issuer, doc["issuer"], "discovery doc must echo the overridden issuer")
}

func TestGRPC_AddRemoveClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithUser(mustUser(t, "u@e.com", "u", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	cl := mustClient(t, "runtime-app",
		dex.WithClientSecret("s"),
		dex.WithClientRedirectURIs("http://localhost/cb"),
		dex.WithClientName("Runtime App"),
	)
	require.NoError(t, c.AddClient(ctx, cl))

	// Idempotency: second Add returns ErrClientExists.
	err = c.AddClient(ctx, cl)
	require.ErrorIs(t, err, dex.ErrClientExists)

	// Removal succeeds.
	require.NoError(t, c.RemoveClient(ctx, cl.ID()))

	// Second remove errors (not-found).
	err = c.RemoveClient(ctx, cl.ID())
	require.ErrorIs(t, err, dex.ErrClientNotFound)
}

func TestGRPC_AddRemoveUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	u := mustUser(t, "runtime@example.com", "runtime", "p")
	require.NoError(t, c.AddUser(ctx, u))

	// Duplicate add errors.
	err = c.AddUser(ctx, u)
	require.ErrorIs(t, err, dex.ErrUserExists)

	// Removal succeeds.
	require.NoError(t, c.RemoveUser(ctx, u.Email()))

	// Second removal errors.
	err = c.RemoveUser(ctx, u.Email())
	require.ErrorIs(t, err, dex.ErrUserNotFound)
}

func TestWithLogger_CapturesDexOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var buf safeBuffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, err := dex.Run(ctx, dexImage,
		dex.WithLogger(logger),
		dex.WithUser(mustUser(t, "u@e.com", "u", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	// Poll for the "listening on" line that Dex always emits at the end of
	// its startup sequence. This proves the log pipe is fully wired — not
	// just that boot started, but that log output flowed through after the
	// container was declared ready.
	//
	// NOTE: Dex v2.x does not log per-operation gRPC events (CreateClient,
	// CreatePassword, etc.) — only boot and key-rotation events reach the
	// log stream. "listening on" is the last boot-time line and therefore
	// the strongest stable signal we can assert on.
	deadline := time.Now().Add(30 * time.Second)
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
		dex.WithClient(mustClient(t, "e2e-app",
			dex.WithClientSecret("e2e-secret"),
			dex.WithClientRedirectURIs(redirectURI),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
			dex.WithClientName("E2E App"),
		)),
		dex.WithUser(mustUser(t, "alice@example.com", "alice", "pass")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

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
	require.NotEmpty(t, tok.AccessToken)

	idToken, ok := tok.Extra("id_token").(string)
	require.True(t, ok, "id_token missing from response")
	require.NotEmpty(t, idToken)
}

func TestAuthCode_RefreshToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const redirectURI = "http://localhost:18080/cb"

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(mustClient(t, "e2e",
			dex.WithClientSecret("s"),
			dex.WithClientName("E2E"),
			dex.WithClientRedirectURIs(redirectURI),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		)),
		dex.WithUser(mustUser(t, "a@e.com", "a", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

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
	require.NotEqual(t, tok.AccessToken, newTok.AccessToken, "refresh yields new access token")
}

func TestAuthCode_MultipleRedirectURIs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	uris := []string{"http://localhost:18080/cb", "http://localhost:18090/cb"}

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(mustClient(t, "e2e",
			dex.WithClientSecret("s"),
			dex.WithClientName("E2E"),
			dex.WithClientRedirectURIs(uris...),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		)),
		dex.WithUser(mustUser(t, "a@e.com", "a", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	for _, uri := range uris {
		t.Run(uri, func(t *testing.T) {
			cfg := oauth2.Config{
				ClientID:     "e2e",
				ClientSecret: "s",
				RedirectURL:  uri,
				Endpoint:     oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
				Scopes:       []string{"openid"},
			}
			tok := drivePasswordAuthCode(t, ctx, cfg, "a@e.com", "p")
			require.NotEmpty(t, tok.AccessToken)
		})
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
		dex.WithClient(mustClient(t, "svc",
			dex.WithClientSecret("s"),
			dex.WithClientName("Service"),
			dex.WithClientGrantTypes("client_credentials"),
		)),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

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
	require.True(t,
		strings.Contains(msg, "unsupported") || strings.Contains(msg, "grant"),
		"expected grant-related error, got: %v", err)
}

func TestPasswordGrant_ROPC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(mustClient(t, "cli",
			dex.WithClientSecret("s"),
			dex.WithClientName("CLI"),
			dex.WithClientGrantTypes("password"),
		)),
		dex.WithUser(mustUser(t, "a@e.com", "a", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	cfg := oauth2.Config{
		ClientID:     "cli",
		ClientSecret: "s",
		Endpoint:     oauth2.Endpoint{TokenURL: c.TokenEndpoint()},
		Scopes:       []string{"openid"},
	}
	tok, err := cfg.PasswordCredentialsToken(ctx, "a@e.com", "p")
	require.NoError(t, err)
	require.NotEmpty(t, tok.AccessToken)
}

func TestMultipleClients_OneInstance(t *testing.T) {
	// Registers two clients on a single Dex instance and verifies both can
	// independently obtain tokens. The "svc" client uses the password grant
	// (Dex's M2M pattern — see TestClientCredentials); the "web" client uses
	// the authorization_code flow. Tokens from distinct clients must differ.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(mustClient(t, "svc",
			dex.WithClientSecret("s"),
			dex.WithClientName("SVC"),
			dex.WithClientGrantTypes("password"),
		)),
		dex.WithClient(mustClient(t, "web",
			dex.WithClientSecret("s"),
			dex.WithClientName("Web"),
			dex.WithClientRedirectURIs("http://localhost/cb"),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		)),
		dex.WithUser(mustUser(t, "a@e.com", "a", "p")),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

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
	require.NotEmpty(t, webTok.AccessToken)

	require.NotEqual(t, svcTok.AccessToken, webTok.AccessToken, "tokens from distinct clients must differ")
}

func TestMockConnector_IssuesToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage,
		dex.WithConnector(dex.ConnectorMock, "mock", "Mock Connector"),
		dex.WithClient(mustClient(t, "e2e",
			dex.WithClientSecret("s"),
			dex.WithClientName("E2E"),
			dex.WithClientRedirectURIs("http://localhost/cb"),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		)),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	cfg := oauth2.Config{
		ClientID: "e2e", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}

	// Drive the /auth URL with connector_id=mock so Dex skips the login form.
	authURL := cfg.AuthCodeURL("state-mock") + "&connector_id=mock"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if strings.HasPrefix(req.URL.String(), cfg.RedirectURL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// CheckRedirect returns http.ErrUseLastResponse on the redirect to
	// cfg.RedirectURL, so resp is always a 3xx with the auth code in
	// its Location header on the success path.
	require.GreaterOrEqual(t, resp.StatusCode, 300, "expected redirect to %s, got status %d", cfg.RedirectURL, resp.StatusCode)
	require.Less(t, resp.StatusCode, 400, "expected redirect to %s, got status %d", cfg.RedirectURL, resp.StatusCode)
	loc, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	code := loc.Query().Get("code")
	require.NotEmpty(t, code, "mockCallback should redirect with ?code=...; got %q", loc.String())

	tok, err := cfg.Exchange(ctx, code)
	require.NoError(t, err)
	require.NotEmpty(t, tok.AccessToken)
}

func TestGRPC_RuntimeAddUsableEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, c)

	// Seed client + user at runtime.
	require.NoError(t, c.AddClient(ctx, mustClient(t, "late-app",
		dex.WithClientSecret("s"),
		dex.WithClientRedirectURIs("http://localhost/cb"),
		dex.WithClientName("Late App"),
	)))
	require.NoError(t, c.AddUser(ctx, mustUser(t, "late@e.com", "late", "p")))

	cfg := oauth2.Config{
		ClientID: "late-app", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}
	tok := drivePasswordAuthCode(t, ctx, cfg, "late@e.com", "p")
	require.NotEmpty(t, tok.AccessToken)

	// Remove user — subsequent login attempt must fail.
	require.NoError(t, c.RemoveUser(ctx, "late@e.com"))

	// Do the login dance manually — drivePasswordAuthCode uses require.NoError
	// which would abort the test on the expected failure.
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if strings.HasPrefix(req.URL.String(), cfg.RedirectURL) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	authURL := cfg.AuthCodeURL("s1")
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	require.NoError(t, err)
	resp, err := client.Do(getReq)
	require.NoError(t, err)
	resp.Body.Close()
	loginURL := resp.Request.URL.String()

	// Negative-path login: POST blindly to the GET's final URL. Dex's
	// local login endpoint accepts POSTs at the same URL the GET returned;
	// the helper's form-action extraction is unnecessary here because we
	// only care that the POST fails, not which rendered path it takes.
	form := url.Values{"login": {"late@e.com"}, "password": {"p"}}
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	require.NoError(t, err)
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2, err := client.Do(postReq)
	require.NoError(t, err)
	body2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()
	// Dex renders the login page again with a failure marker. The exact
	// copy ("Invalid Email Address and password") may drift across
	// versions; match loosely.
	lower := strings.ToLower(string(body2))
	require.True(t,
		strings.Contains(lower, "invalid") || strings.Contains(lower, "authentication failed") || r2.StatusCode >= 400,
		"removed user should fail login; got status=%d body-prefix=%.200q", r2.StatusCode, lower)
}

func TestWithIssuer_CrossContainerViaNetworkAlias(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	net, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

	const issuer = "http://dex:5556/dex"

	c, err := dex.Run(ctx, dexImage,
		dex.WithIssuer(issuer),
		dex.WithUser(mustUser(t, "u@e.com", "u", "p")),
		network.WithNetwork([]string{"dex"}, net),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, c)

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
	testcontainers.CleanupContainer(t, sidecar)

	logs, err := sidecar.Logs(ctx)
	require.NoError(t, err)
	defer logs.Close()
	body, err := io.ReadAll(logs)
	require.NoError(t, err)

	require.Contains(t, string(body), issuer,
		"discovery doc fetched via network alias must echo overridden issuer")
}

func TestClientCredentials_WithFeatureFlag(t *testing.T) {
	// dexImageWithCC is the floating dexidp/dex:master tag. Its contents
	// shift without warning, so this test opts in via DEX_TEST_MASTER=1 to
	// keep the default CI run deterministic. Once Dex v2.46.0 ships, swap
	// dexImageWithCC for the pinned tag and drop this gate.
	if os.Getenv("DEX_TEST_MASTER") != "1" {
		t.Skip("set DEX_TEST_MASTER=1 to run; uses floating dexidp/dex:master tag")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImageWithCC,
		dex.WithEnableClientCredentials(),
		dex.WithClient(mustClient(t, "svc",
			dex.WithClientSecret("s"),
			dex.WithClientName("Service"),
			dex.WithClientGrantTypes("client_credentials"),
		)),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, c)

	cfg := clientcredentials.Config{
		ClientID:     "svc",
		ClientSecret: "s",
		TokenURL:     c.TokenEndpoint(),
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	tok, err := cfg.TokenSource(ctx).Token()
	require.NoError(t, err, "CC grant should succeed with feature flag enabled on master image")
	require.NotEmpty(t, tok.AccessToken)
}

func TestConsumer_IDTokenVerifies_CoreosOIDC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const redirectURI = "http://localhost:18080/cb"

	c, err := dex.Run(ctx, dexImage,
		dex.WithClient(mustClient(t, "e2e",
			dex.WithClientSecret("s"),
			dex.WithClientName("E2E"),
			dex.WithClientRedirectURIs(redirectURI),
			dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		)),
		dex.WithUser(mustUser(t, "a@e.com", "a", "p")),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, c)

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
	require.Equal(t, "a@e.com", claims.Email)
	require.Equal(t, "a", claims.Name)
	require.NotEmpty(t, claims.Sub)
}
