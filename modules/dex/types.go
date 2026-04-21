package dex

import "errors"

// Client is a static OAuth2 client registered with Dex.
type Client struct {
	// ID is the client_id. Required.
	ID string
	// Secret is the client secret. Required unless Public is true.
	Secret string
	// RedirectURIs lists allowed redirect URIs. At least one is required for
	// authorization_code clients.
	RedirectURIs []string
	// GrantTypes lists OAuth2 grants the client may use. Defaults to
	// ["authorization_code", "refresh_token"]. Values: authorization_code,
	// refresh_token, client_credentials, password.
	//
	// Only takes effect for clients registered via WithClient (YAML).
	// Clients added at runtime via AddClient inherit Dex's defaults
	// (authorization_code + refresh_token) because the gRPC api.Client proto
	// has no grant_types field.
	GrantTypes []string
	// Public marks the client as public (no secret). Used for PKCE flows.
	Public bool
	// Name is the human-readable display name shown on Dex's consent screen.
	Name string
}

// User is a static password entry in Dex's password connector.
type User struct {
	// Email is the user's email address. Required.
	Email string
	// Username is the preferred_username claim. Required.
	Username string
	// Password is the cleartext password. Bcrypted internally. Required.
	Password string
	// UserID is the stable subject claim. If empty, a UUID is generated.
	UserID string
}

// ConnectorType selects a Dex connector kind.
type ConnectorType string

const (
	// ConnectorPassword enables Dex's built-in static password connector.
	// Users must be registered via WithUser or AddUser.
	ConnectorPassword ConnectorType = "password"
	// ConnectorMock enables Dex's mockCallback connector — a test-only
	// connector that bypasses the login form and returns a fixed user.
	ConnectorMock ConnectorType = "mockCallback"
)

var (
	// ErrClientExists is returned by AddClient when a client with the given
	// ID is already registered.
	ErrClientExists = errors.New("dex: client already exists")
	// ErrUserExists is returned by AddUser when a user with the given email
	// is already registered.
	ErrUserExists = errors.New("dex: user already exists")
	// ErrNoAuthSource is returned when the rendered Dex config would boot
	// with no working authentication source — neither the password DB nor
	// any connector. The password DB is enabled by default.
	ErrNoAuthSource = errors.New("dex: no auth source configured")
)
