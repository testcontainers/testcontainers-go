package dex

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	dexapi "github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	HTTPPort = "5556/tcp"
	GRPCPort = "5557/tcp"
)

const (
	bcryptCost    = 10
	defaultIssuer = "http://localhost:5556"
)

var ErrPasswordRequired = errors.New("plaintext password or hash is required")

var (
	propertiesToPatch = []string{
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"userinfo_endpoint",
		"device_authorization_endpoint",
		"introspection_endpoint",
	}
	//go:embed config/dex-config.yaml.tmpl
	dexConfigTemplate []byte
)

// CreateClientAppRequest represents a request to register a client application in Dex.
type CreateClientAppRequest struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURIs []string
	Public       bool
}

// CreatePasswordRequest represents a request to register an identity in Dex.
// Either Hash or Password must be provided.
// If Hash is provided, Password will be ignored.
type CreatePasswordRequest struct {
	UserID   string
	Email    string
	Username string
	// Password is a plain text password.
	Password string
	// Hash is a bcrypt hash of the password.
	Hash []byte
}

// Container represents the Dex container type used in the module
type Container struct {
	testcontainers.Container
	Client *http.Client
}

// CreateClientApp registers a new client application in Dex.
// If ClientID or ClientSecret are empty, they will be generated and returned.
func (c Container) CreateClientApp(ctx context.Context, req CreateClientAppRequest) (*dexapi.Client, error) {
	apiClient, connCloser, err := c.createDexAPIClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare Dex API client: %w", err)
	}

	defer func() {
		if closeErr := connCloser.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close GRPC connection: %w", closeErr))
		}
	}()

	if req.ClientID == "" {
		req.ClientID = uuid.New().String()
	}

	if req.ClientSecret == "" {
		req.ClientSecret = randomSecret()
	}

	apiReq := &dexapi.CreateClientReq{
		Client: &dexapi.Client{
			Name:         req.Name,
			Id:           req.ClientID,
			Secret:       req.ClientSecret,
			RedirectUris: req.RedirectURIs,
			Public:       req.Public,
		},
	}

	resp, err := apiClient.CreateClient(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("register client application: %w", err)
	}

	return resp.Client, nil
}

// CreatePassword registers a new 'password' (login) in Dex. Either Password or Hash must be provided.
// If Hash is provided, Password will be ignored. If UserID is not set, a UUID will be generated.
// The password will be hashed using bcrypt if provided. Returns ErrPasswordRequired if neither
// Password nor Hash are provided, or an error from bcrypt if hashing fails.
func (c Container) CreatePassword(
	ctx context.Context,
	req CreatePasswordRequest,
) (err error) {
	apiClient, connCloser, err := c.createDexAPIClient(ctx)
	if err != nil {
		return fmt.Errorf("prepare Dex API client: %w", err)
	}

	defer func() {
		if closeErr := connCloser.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close GRPC connection: %w", closeErr))
		}
	}()

	if req.Password == "" && len(req.Hash) == 0 {
		return ErrPasswordRequired
	}

	if req.UserID == "" {
		req.UserID = uuid.New().String()
	}

	if len(req.Hash) == 0 {
		req.Hash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
		if err != nil {
			return fmt.Errorf("hash plaintext password with bcrypt: %w", err)
		}
	}

	apiReq := &dexapi.CreatePasswordReq{
		Password: &dexapi.Password{
			UserId:   req.UserID,
			Username: req.Username,
			Email:    req.Email,
			Hash:     req.Hash,
		},
	}

	if _, err = apiClient.CreatePassword(ctx, apiReq); err != nil {
		return fmt.Errorf("create password in Dex: %w", err)
	}

	return nil
}

// OpenIDConfiguration returns the OpenID configuration for the Dex instance.
// It retrieves the raw configuration, unmarshals it into an OpenIDConfiguration struct,
// and returns any error that occurs during the process.
func (c Container) OpenIDConfiguration(ctx context.Context) (cfg OpenIDConfiguration, err error) {
	rawCfg, err := c.RawOpenIDConfiguration(ctx)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("unmarshal OpenID configuration: %w", err)
	}

	return cfg, nil
}

// RawOpenIDConfiguration retrieves the raw OpenID configuration from the Dex instance.
// It fetches the configuration from the .well-known/openid-configuration endpoint,
// patches all endpoint URLs to use the container's HTTP endpoint, and returns the
// modified configuration as JSON bytes. Returns an error if the request fails
// or if parsing/marshaling the configuration fails.
func (c Container) RawOpenIDConfiguration(ctx context.Context) (rawCfg []byte, err error) {
	httpEndpoint, err := c.PortEndpoint(ctx, HTTPPort, "http")
	if err != nil {
		return nil, fmt.Errorf("get Dex HTTP endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpEndpoint+"/.well-known/openid-configuration", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("prepare OIDC discovery requrest: %w", err)
	}

	httpClient := c.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send OIDC discovery request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var partiallyDecoded map[string]json.RawMessage

	configDecoder := json.NewDecoder(resp.Body)
	if err := configDecoder.Decode(&partiallyDecoded); err != nil {
		return nil, fmt.Errorf("decode OIDC discovery response: %w", err)
	}

	for _, property := range propertiesToPatch {
		jsonValue, ok := partiallyDecoded[property]
		if !ok {
			continue
		}

		patched, err := patchEndpoint(strings.Trim(string(jsonValue), "\""), httpEndpoint)
		if err != nil {
			return nil, err
		}
		partiallyDecoded[property] = json.RawMessage(fmt.Sprintf("%q", patched))
	}

	if rawCfg, err = json.Marshal(partiallyDecoded); err != nil {
		return nil, fmt.Errorf("marshal OIDC discovery response: %w", err)
	}

	return rawCfg, nil
}

func (c Container) createDexAPIClient(ctx context.Context) (dexapi.DexClient, io.Closer, error) {
	grpcEndpoint, err := c.PortEndpoint(ctx, GRPCPort, "")
	if err != nil {
		return nil, nil, fmt.Errorf("get GRPC port endpoint: %w", err)
	}

	grpcClient, err := grpc.NewClient(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("create Dex API gRPC client: %w", err)
	}

	return dexapi.NewDexClient(grpcClient), grpcClient, nil
}

// Run creates an instance of the Dex container type.
// The image does not have a default value, it **must** be provided.
func Run(
	ctx context.Context,
	img string,
	opts ...testcontainers.ContainerCustomizer,
) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Env: map[string]string{
			"DEX_ISSUER": defaultIssuer,
		},
		ExposedPorts: []string{
			HTTPPort,
			GRPCPort,
		},
		WaitingFor: wait.ForAll(
			wait.ForMappedPort(GRPCPort),
			wait.ForHTTP("/healthz").WithPort(HTTPPort),
		),
		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewReader(dexConfigTemplate),
				ContainerFilePath: "/etc/dex/config.docker.yaml",
				FileMode:          0o644,
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize container request: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func patchEndpoint(original, newHost string) (patched string, err error) {
	if original == "" {
		return "", nil
	}

	parsedOriginalURL, err := url.Parse(original)
	if err != nil {
		return "", fmt.Errorf("parse original URL %s: %w", original, err)
	}

	parsedPatchedURL, err := url.Parse(newHost)
	if err != nil {
		return "", fmt.Errorf("parse container HTTP port endpoint %s: %w", newHost, err)
	}

	parsedOriginalURL.Scheme = parsedPatchedURL.Scheme
	parsedOriginalURL.Host = parsedPatchedURL.Host

	patched = parsedOriginalURL.String()
	return patched, nil
}

// randomSecret generates a random password for identities.
// Based on https://pkg.go.dev/crypto/rand@go1.24.0#Text
// Can be replaced as soon as testcontainers-go is updated to Go 1.24 or higher.
func randomSecret() string {
	// ⌈log₃₂ 2¹²⁸⌉ = 26 chars
	const (
		length         = 26
		base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	)

	src := make([]byte, length)
	_, _ = rand.Read(src)
	for i := range src {
		src[i] = base32alphabet[src[i]%32]
	}
	return string(src)
}
