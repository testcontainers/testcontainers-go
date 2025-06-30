package dex

import (
	"errors"

	dexapi "github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/testcontainers/testcontainers-go"
)

var errEmptyUserID = errors.New("user ID cannot be empty")

// WithIssuer sets the issuer for the Dex container.
// In most cases, the issuer should be set to the URL of the Dex container.
// If not provided, the default issuer will be used, that will not necessarily match the actual URL of the Dex container.
func WithIssuer(issuer string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DEX_ISSUER"] = issuer
		return nil
	})
}

// WithLogLevel overrides the default log level for the Dex container.
// See [log/slog](log/slog#Level) for possible values.
func WithLogLevel(level string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DEX_LOG_LEVEL"] = level
		return nil
	})
}

type CreatePasswordCredential func() (email string, passwordHash []byte, err error)

func PlainTextCredential(email, password string) CreatePasswordCredential {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return func() (string, []byte, error) {
		return email, hashedPassword, err
	}
}

func HashedCredential(email string, passwordHash []byte) CreatePasswordCredential {
	return func() (string, []byte, error) {
		return email, passwordHash, nil
	}
}

func combineOptionsToPassword(email string, hash []byte, opts []CreatePasswordOption) (*dexapi.Password, error) {
	options := createPasswordOptions{
		UserID: uuid.New().String(),
	}

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	return &dexapi.Password{
		Email:    email,
		Hash:     hash,
		Username: options.Username,
		UserId:   options.UserID,
	}, nil
}

type createPasswordOptions struct {
	Username string
	UserID   string
}

type CreatePasswordOption func(pw *createPasswordOptions) error

func WithUserID(userID string) CreatePasswordOption {
	return func(pw *createPasswordOptions) error {
		if userID == "" {
			return errEmptyUserID
		}

		pw.UserID = userID
		return nil
	}
}

func WithUsername(username string) CreatePasswordOption {
	return func(pw *createPasswordOptions) error {
		pw.Username = username
		return nil
	}
}

func combineOptionsToClientApp(opts []CreateClientAppOption) (*dexapi.Client, error) {
	var options createClientAppOptions

	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, err
		}
	}

	return &dexapi.Client{
		Id:           options.ID,
		Secret:       options.Secret,
		RedirectUris: options.RedirectUris,
		TrustedPeers: options.TrustedPeers,
		Public:       options.Public,
		Name:         options.Name,
		LogoUrl:      options.LogoURL,
	}, nil
}

type createClientAppOptions struct {
	ID           string
	Secret       string
	RedirectUris []string
	TrustedPeers []string
	Public       bool
	Name         string
	LogoURL      string
}

type CreateClientAppOption func(ca *createClientAppOptions) error

func WithClientName(clientName string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.Name = clientName
		return nil
	}
}

func WithClientID(clientID string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.ID = clientID
		return nil
	}
}

func WithClientSecret(clientSecret string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.Secret = clientSecret
		return nil
	}
}

func WithRedirectURIs(redirectURIs ...string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.RedirectUris = redirectURIs
		return nil
	}
}

func WithPublicClient(public bool) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.Public = public
		return nil
	}
}

func WithTrustedPeers(trustedPeers ...string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.TrustedPeers = trustedPeers
		return nil
	}
}

func WithLogoURL(logoURL string) CreateClientAppOption {
	return func(ca *createClientAppOptions) error {
		ca.LogoURL = logoURL
		return nil
	}
}
