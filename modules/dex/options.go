package dex

import (
	dexapi "github.com/dexidp/dex/api/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/testcontainers/testcontainers-go"
)

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

type CreatePasswordOption func(pw *dexapi.Password)

func WithUserID(userID string) CreatePasswordOption {
	return func(pw *dexapi.Password) {
		pw.UserId = userID
	}
}

func WithUsername(username string) CreatePasswordOption {
	return func(pw *dexapi.Password) {
		pw.Username = username
	}
}

type CreateClientAppOption func(ca *dexapi.Client)

func WithClientName(clientName string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.Name = clientName
	}
}

func WithClientID(clientID string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.Id = clientID
	}
}

func WithClientSecret(clientSecret string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.Secret = clientSecret
	}
}

func WithRedirectURIs(redirectURIs []string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.RedirectUris = redirectURIs
	}
}

func WithPublicClient(public bool) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.Public = public
	}
}

func WithTrustedPeers(trustedPeers []string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.TrustedPeers = trustedPeers
	}
}

func WithLogoURL(logoURL string) CreateClientAppOption {
	return func(ca *dexapi.Client) {
		ca.LogoUrl = logoURL
	}
}
