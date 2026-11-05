package dexexamples_test

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
	dexexamples "github.com/testcontainers/testcontainers-go/modules/dex/examples"
)

const (
	dexImage       = "dexidp/dex:v2.45.1"
	dexImageWithCC = "dexidp/dex:master"
)

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
	require.NoError(t, dexexamples.NewAdminClient(c).AddClient(ctx, cl))

	// Idempotency: second Add returns ErrClientExists.
	err = dexexamples.NewAdminClient(c).AddClient(ctx, cl)
	require.ErrorIs(t, err, dex.ErrClientExists)

	// Removal succeeds.
	require.NoError(t, dexexamples.NewAdminClient(c).RemoveClient(ctx, cl.ID()))

	// Second remove errors (not-found).
	err = dexexamples.NewAdminClient(c).RemoveClient(ctx, cl.ID())
	require.ErrorIs(t, err, dex.ErrClientNotFound)
}

func TestGRPC_AddRemoveUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	u := mustUser(t, "runtime@example.com", "runtime", "p")
	require.NoError(t, dexexamples.NewAdminClient(c).AddUser(ctx, u))

	// Duplicate add errors.
	err = dexexamples.NewAdminClient(c).AddUser(ctx, u)
	require.ErrorIs(t, err, dex.ErrUserExists)

	// Removal succeeds.
	require.NoError(t, dexexamples.NewAdminClient(c).RemoveUser(ctx, u.Email()))

	// Second removal errors.
	err = dexexamples.NewAdminClient(c).RemoveUser(ctx, u.Email())
	require.ErrorIs(t, err, dex.ErrUserNotFound)
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
	require.NoError(t, dexexamples.NewAdminClient(c).AddClient(ctx, mustClient(t, "late-app",
		dex.WithClientSecret("s"),
		dex.WithClientRedirectURIs("http://localhost/cb"),
		dex.WithClientName("Late App"),
	)))
	require.NoError(t, dexexamples.NewAdminClient(c).AddUser(ctx, mustUser(t, "late@e.com", "late", "p")))

	cfg := oauth2.Config{
		ClientID: "late-app", ClientSecret: "s",
		RedirectURL: "http://localhost/cb",
		Endpoint:    oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:      []string{"openid"},
	}
	tok := drivePasswordAuthCode(t, ctx, cfg, "late@e.com", "p")
	require.NotEmpty(t, tok.AccessToken)

	// Remove user — subsequent login attempt must fail.
	require.NoError(t, dexexamples.NewAdminClient(c).RemoveUser(ctx, "late@e.com"))

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
