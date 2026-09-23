package dex

import (
	"errors"
	"fmt"
)

// validClientGrantTypes is the set of OAuth2 grant types Dex understands.
// Kept in sync with WithClientGrantTypes' godoc.
var validClientGrantTypes = map[string]struct{}{
	"authorization_code": {},
	"refresh_token":      {},
	"client_credentials": {},
	"password":           {},
}

// Client is a static OAuth2 client registered with Dex at boot time via
// WithClient. Construct with NewClient so invalid configuration surfaces at
// call-site rather than at Run.
type Client struct {
	id           string
	secret       string
	name         string
	redirectURIs []string
	grantTypes   []string
	public       bool
}

// ID returns the client_id.
func (c Client) ID() string { return c.id }

// ClientOption configures a Client during NewClient.
type ClientOption func(*Client) error

// NewClient creates a Client registered statically at boot. ID is required;
// every other field is optional and may be set through ClientOption values.
//
// Returns an error when the ID is blank or any option rejects its input.
func NewClient(id string, opts ...ClientOption) (Client, error) {
	if id == "" {
		return Client{}, errors.New("dex: client id must not be blank")
	}
	c := Client{id: id}
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return Client{}, err
		}
	}
	return c, nil
}

// WithClientSecret sets the client secret. Required for confidential clients;
// omit for public (PKCE) clients via WithClientPublic.
func WithClientSecret(s string) ClientOption {
	return func(c *Client) error {
		if s == "" {
			return errors.New("dex: client secret must not be blank")
		}
		c.secret = s
		return nil
	}
}

// WithClientName sets the human-readable display name shown on Dex's consent
// screen.
func WithClientName(n string) ClientOption {
	return func(c *Client) error {
		if n == "" {
			return errors.New("dex: client name must not be blank")
		}
		c.name = n
		return nil
	}
}

// WithClientRedirectURIs appends to the list of allowed redirect URIs. At
// least one is required for authorization_code clients. Values are appended
// across calls; blank entries are rejected.
func WithClientRedirectURIs(uris ...string) ClientOption {
	return func(c *Client) error {
		for _, u := range uris {
			if u == "" {
				return errors.New("dex: client redirect URI must not be blank")
			}
		}
		c.redirectURIs = append(c.redirectURIs, uris...)
		return nil
	}
}

// WithClientGrantTypes appends to the allowed OAuth2 grants. Defaults to
// ["authorization_code", "refresh_token"] when unset. Accepted values:
// authorization_code, refresh_token, client_credentials, password.
//
// Only takes effect for clients registered via WithClient (YAML). Clients
// added at runtime via AddClient inherit Dex's defaults because the gRPC
// api.Client proto has no grant_types field.
func WithClientGrantTypes(grants ...string) ClientOption {
	return func(c *Client) error {
		for _, g := range grants {
			if g == "" {
				return errors.New("dex: client grant type must not be blank")
			}
			if _, ok := validClientGrantTypes[g]; !ok {
				return fmt.Errorf("dex: unsupported client grant type %q", g)
			}
		}
		c.grantTypes = append(c.grantTypes, grants...)
		return nil
	}
}

// WithClientPublic marks the client as public — no secret, intended for PKCE
// flows from untrusted clients (mobile, SPA).
func WithClientPublic() ClientOption {
	return func(c *Client) error {
		c.public = true
		return nil
	}
}

// User is a static password entry in Dex's password connector. Construct
// with NewUser.
type User struct {
	email    string
	username string
	password string
	userID   string
}

// Email returns the email address.
func (u User) Email() string { return u.email }

// UserOption configures a User during NewUser.
type UserOption func(*User) error

// NewUser creates a static password entry. Email, username and password are
// required; a user ID may be pinned via WithUserID (else a UUIDv4 is
// generated at YAML render time).
func NewUser(email, username, password string, opts ...UserOption) (User, error) {
	if email == "" {
		return User{}, errors.New("dex: user email must not be blank")
	}
	if username == "" {
		return User{}, errors.New("dex: user username must not be blank")
	}
	if password == "" {
		return User{}, errors.New("dex: user password must not be blank")
	}
	u := User{email: email, username: username, password: password}
	for _, opt := range opts {
		if err := opt(&u); err != nil {
			return User{}, err
		}
	}
	return u, nil
}

// WithUserID pins the stable subject claim. When unset, NewUser leaves
// userID blank and a UUIDv4 is generated at YAML render time.
func WithUserID(id string) UserOption {
	return func(u *User) error {
		if id == "" {
			return errors.New("dex: user id must not be blank")
		}
		u.userID = id
		return nil
	}
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

// Storage selects Dex's storage backend. Defaults to StorageSQLite.
type Storage string

const (
	// StorageSQLite uses an on-disk SQLite database inside the container.
	// Ephemeral — destroyed when the container is removed.
	StorageSQLite Storage = "sqlite3"
	// StorageMemory keeps all state in process memory. Fastest; unsuitable
	// when multiple Dex replicas need to share state.
	StorageMemory Storage = "memory"
)

var (
	// ErrClientExists is returned by AddClient when a client with the given
	// ID is already registered.
	ErrClientExists = errors.New("dex: client already exists")
	// ErrClientNotFound is returned by RemoveClient when no client matches
	// the given ID.
	ErrClientNotFound = errors.New("dex: client not found")
	// ErrUserExists is returned by AddUser when a user with the given email
	// is already registered.
	ErrUserExists = errors.New("dex: user already exists")
	// ErrUserNotFound is returned by RemoveUser when no user matches the
	// given email.
	ErrUserNotFound = errors.New("dex: user not found")
	// ErrNoAuthSource is returned when the rendered Dex config would boot
	// with no working authentication source — neither the password DB nor
	// any connector. The password DB is enabled by default; callers must
	// explicitly disable it via WithDisablePasswordDB to trigger this error.
	ErrNoAuthSource = errors.New("dex: no auth source configured")
)
