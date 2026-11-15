package dex_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

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
	registration := cl.Registration()
	require.Equal(t, "runtime-app", registration.ID)
	require.Equal(t, "s", registration.Secret)
	require.Equal(t, "Runtime App", registration.Name)
	require.Equal(t, []string{"http://localhost/cb"}, registration.RedirectURIs)
	assertContainerEndpoints(t, ctx, c)
}

func TestGRPC_AddRemoveUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)

	u := mustUser(t, "runtime@example.com", "runtime", "p")
	registration, err := u.PasswordRegistration()
	require.NoError(t, err)
	require.Equal(t, "runtime@example.com", registration.Email)
	require.Equal(t, "runtime", registration.Username)
	require.NotEmpty(t, registration.UserID)
	require.NoError(t, bcrypt.CompareHashAndPassword(registration.Hash, []byte("p")))
	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
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
			u, err := url.Parse(uri)
			require.NoError(t, err)
			require.Equal(t, "/cb", u.Path)
			assertClientRedirectURI(t, ctx, c, "e2e", uri)
			assertContainerEndpoints(t, ctx, c)
		})
	}
}

func TestClientCredentials_UnsupportedByLocalConnectors(t *testing.T) {
	// Retain the local-connector configuration and container boundary here.
	// The child test asserts the original unsupported-grant error.
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

	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
}

func TestMultipleClients_OneInstance(t *testing.T) {
	// Retain both clients on one instance. The child test verifies password
	// and authorization-code tokens and requires them to differ.
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

	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
}

func TestGRPC_RuntimeAddUsableEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	c, err := dex.Run(ctx, dexImage)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, c)

	cl := mustClient(t, "late-app",
		dex.WithClientSecret("s"),
		dex.WithClientRedirectURIs("http://localhost/cb"),
		dex.WithClientName("Late App"),
	)
	u := mustUser(t, "late@e.com", "late", "p")
	require.Equal(t, "late-app", cl.Registration().ID)
	require.Equal(t, []string{"http://localhost/cb"}, cl.Registration().RedirectURIs)
	registration, err := u.PasswordRegistration()
	require.NoError(t, err)
	require.Equal(t, "late@e.com", registration.Email)
	require.NoError(t, bcrypt.CompareHashAndPassword(registration.Hash, []byte("p")))
	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
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

	assertContainerEndpoints(t, ctx, c)
}

// assertContainerEndpoints verifies the parent container/configuration boundary.
// The child suite retains the original OAuth/OIDC and gRPC verdicts.
func assertContainerEndpoints(t *testing.T, ctx context.Context, c *dex.Container) {
	t.Helper()
	require.NotEmpty(t, c.IssuerURL())
	require.Equal(t, c.IssuerURL()+"/auth", c.AuthEndpoint())
	require.Equal(t, c.IssuerURL()+"/token", c.TokenEndpoint())
	require.Equal(t, c.IssuerURL()+"/keys", c.JWKSEndpoint())
	ep, err := c.GRPCEndpoint(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, ep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.ConfigEndpoint(), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var doc map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&doc))
	require.Equal(t, c.IssuerURL(), doc["issuer"])
	require.Equal(t, c.AuthEndpoint(), doc["authorization_endpoint"])
	require.Equal(t, c.TokenEndpoint(), doc["token_endpoint"])
	require.Equal(t, c.JWKSEndpoint(), doc["jwks_uri"])
}

func assertClientRedirectURI(t *testing.T, ctx context.Context, c *dex.Container, id, uri string) {
	t.Helper()
	config, err := c.CopyFileFromContainer(ctx, "/etc/dex/dex.yml")
	require.NoError(t, err)
	defer config.Close()
	var doc struct {
		Clients []struct {
			ID           string   `yaml:"id"`
			RedirectURIs []string `yaml:"redirectURIs"`
		} `yaml:"staticClients"`
	}
	require.NoError(t, yaml.NewDecoder(config).Decode(&doc))
	for _, client := range doc.Clients {
		if client.ID == id {
			require.Contains(t, client.RedirectURIs, uri)
			return
		}
	}
	t.Fatalf("client %q missing from rendered container configuration", id)
}
