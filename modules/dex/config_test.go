package dex

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

func TestRender_MinimalDefaults(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://localhost:5556/dex"

	out, err := render(o)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, yaml.Unmarshal(out, &got))

	assert.Equal(t, "http://localhost:5556/dex", got["issuer"])
	assert.Equal(t, true, got["enablePasswordDB"])
	storage := got["storage"].(map[string]any)
	assert.Equal(t, "sqlite3", storage["type"])
	web := got["web"].(map[string]any)
	assert.Equal(t, "0.0.0.0:5556", web["http"])
	grpc := got["grpc"].(map[string]any)
	assert.Equal(t, "0.0.0.0:5557", grpc["addr"])
	oauth2 := got["oauth2"].(map[string]any)
	assert.Equal(t, true, oauth2["skipApprovalScreen"])
}

func TestRender_WithClients(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.clients = []Client{
		{
			ID: "app1", Secret: "s1",
			RedirectURIs: []string{"http://a/cb", "http://b/cb"},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
			Name:         "App 1",
		},
		{
			ID: "svc", Secret: "s2",
			GrantTypes: []string{"client_credentials"},
			Name:       "Service",
		},
	}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticClients []struct {
			ID           string   `yaml:"id"`
			Secret       string   `yaml:"secret"`
			RedirectURIs []string `yaml:"redirectURIs"`
			GrantTypes   []string `yaml:"grantTypes"`
			Name         string   `yaml:"name"`
		} `yaml:"staticClients"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.StaticClients, 2)
	assert.Equal(t, "app1", got.StaticClients[0].ID)
	assert.Equal(t, []string{"http://a/cb", "http://b/cb"}, got.StaticClients[0].RedirectURIs)
	assert.Equal(t, []string{"client_credentials"}, got.StaticClients[1].GrantTypes)
}

func TestRender_WithUsers_BcryptShape(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{{Email: "u@e.com", Username: "u", Password: "p"}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticPasswords []struct {
			Email    string `yaml:"email"`
			Hash     string `yaml:"hash"`
			Username string `yaml:"username"`
			UserID   string `yaml:"userID"`
		} `yaml:"staticPasswords"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.StaticPasswords, 1)
	p := got.StaticPasswords[0]
	assert.Equal(t, "u@e.com", p.Email)
	assert.True(t, strings.HasPrefix(p.Hash, "$2a$") || strings.HasPrefix(p.Hash, "$2b$"), "bcrypt prefix")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(p.Hash), []byte("p")))
	assert.NotEmpty(t, p.UserID, "userID should be auto-populated")
}

func TestRender_WithConnectors(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.connectors = []connector{{Type: ConnectorMock, ID: "mock", Name: "Mock"}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		Connectors []struct {
			Type string `yaml:"type"`
			ID   string `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"connectors"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.Connectors, 1)
	assert.Equal(t, "mockCallback", got.Connectors[0].Type)
}

func TestRender_NoAuthSource_Errors(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.enablePasswordDB = false
	// no connectors

	_, err := render(o)
	assert.ErrorIs(t, err, ErrNoAuthSource)
}

func TestRender_IssuerRequired(t *testing.T) {
	o := defaultOptions()
	// issuer empty
	_, err := render(o)
	assert.Error(t, err, "render must error when issuer is empty")
}

func TestRender_BcryptCost(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{{Email: "u@e.com", Username: "u", Password: "p"}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticPasswords []struct {
			Hash string `yaml:"hash"`
		} `yaml:"staticPasswords"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	cost, err := bcrypt.Cost([]byte(got.StaticPasswords[0].Hash))
	require.NoError(t, err)
	// Dex v2.45+ enforces a minimum bcrypt cost of 10; we use exactly 10 to
	// satisfy that constraint while staying well below the production default (14).
	assert.Equal(t, 10, cost, "bcrypt cost must be exactly 10: meets Dex minimum, fast enough for tests")
}

func TestRender_UserWithExplicitUserID(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{{Email: "u@e.com", Username: "u", Password: "p", UserID: "fixed-id-123"}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticPasswords []struct {
			UserID string `yaml:"userID"`
		} `yaml:"staticPasswords"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.StaticPasswords, 1)
	assert.Equal(t, "fixed-id-123", got.StaticPasswords[0].UserID)
}

func TestRender_PasswordConnector_SetWhenPasswordDBEnabled(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	// enablePasswordDB is true by default.

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		OAuth2 struct {
			PasswordConnector string `yaml:"passwordConnector"`
		} `yaml:"oauth2"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	assert.Equal(t, "local", got.OAuth2.PasswordConnector,
		"oauth2.passwordConnector must be 'local' when enablePasswordDB is true")
}

func TestRender_PasswordConnector_OmitWhenPasswordDBDisabled(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.enablePasswordDB = false
	o.connectors = []connector{{Type: ConnectorMock, ID: "mock", Name: "Mock"}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		OAuth2 struct {
			PasswordConnector string `yaml:"passwordConnector"`
		} `yaml:"oauth2"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	assert.Empty(t, got.OAuth2.PasswordConnector,
		"oauth2.passwordConnector must be omitted when enablePasswordDB is false")
}

func TestRender_YAMLInjection_NameField(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	// A name containing a newline + bogus YAML that would break a
	// template-based renderer. yaml.Marshal must escape it such that
	// the round-tripped value equals the input verbatim.
	malicious := "real-name\nmalicious_key: poison"
	o.clients = []Client{{ID: "c", Secret: "s", Name: malicious}}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticClients []struct {
			Name string `yaml:"name"`
		} `yaml:"staticClients"`
		Malicious string `yaml:"malicious_key,omitempty"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.StaticClients, 1)
	assert.Equal(t, malicious, got.StaticClients[0].Name, "injected characters must round-trip as data, not structure")
	assert.Empty(t, got.Malicious, "structural injection must not create a top-level key")
}

func TestParseImageTag(t *testing.T) {
	cases := map[string]string{
		"dexidp/dex:v2.45.1":                     "v2.45.1",
		"dexidp/dex:master":                      "master",
		"dexidp/dex:latest-alpine":               "latest-alpine",
		"ghcr.io/dexidp/dex:v2.46.0":             "v2.46.0",
		"localhost:5000/dex:v2.46.0":             "v2.46.0",
		"dexidp/dex":                             "",
		"dexidp/dex@sha256:abcd":                 "",
		"dexidp/dex:v2.46.0@sha256:deadbeefcafe": "v2.46.0",
	}
	for in, want := range cases {
		assert.Equal(t, want, parseImageTag(in), "input=%q", in)
	}
}

func TestImageSupportsClientCredentials(t *testing.T) {
	good := []string{"master", "latest", "master-alpine", "latest-distroless",
		"v2.46.0", "v2.46.1", "v2.47.0", "v3.0.0",
		"custom-build-abc", // unknown non-semver → trusted
	}
	for _, tag := range good {
		assert.True(t, imageSupportsClientCredentials(tag), "tag=%q should be supported", tag)
	}

	bad := []string{"v2.45.0", "v2.45.1", "v2.45.1-alpine", "v2.44.0",
		"v2.0.0", "v1.99.99"}
	for _, tag := range bad {
		assert.False(t, imageSupportsClientCredentials(tag), "tag=%q should NOT be supported", tag)
	}
}
