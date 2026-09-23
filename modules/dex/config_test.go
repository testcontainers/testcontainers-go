package dex

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

func mustClient(t *testing.T, id string, opts ...ClientOption) Client {
	t.Helper()
	c, err := NewClient(id, opts...)
	require.NoError(t, err)
	return c
}

func mustUser(t *testing.T, email, username, password string, opts ...UserOption) User {
	t.Helper()
	u, err := NewUser(email, username, password, opts...)
	require.NoError(t, err)
	return u
}

func TestRender_MinimalDefaults(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://localhost:5556/dex"

	out, err := render(o)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, yaml.Unmarshal(out, &got))

	require.Equal(t, "http://localhost:5556/dex", got["issuer"])
	require.Equal(t, true, got["enablePasswordDB"])
	storage := got["storage"].(map[string]any)
	require.Equal(t, "sqlite3", storage["type"])
	web := got["web"].(map[string]any)
	require.Equal(t, "0.0.0.0:5556", web["http"])
	grpc := got["grpc"].(map[string]any)
	require.Equal(t, "0.0.0.0:5557", grpc["addr"])
	oauth2 := got["oauth2"].(map[string]any)
	require.Equal(t, true, oauth2["skipApprovalScreen"])
}

func TestRender_WithClients(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.clients = []Client{
		mustClient(t, "app1",
			WithClientSecret("s1"),
			WithClientRedirectURIs("http://a/cb", "http://b/cb"),
			WithClientGrantTypes("authorization_code", "refresh_token"),
			WithClientName("App 1"),
		),
		mustClient(t, "svc",
			WithClientSecret("s2"),
			WithClientGrantTypes("client_credentials"),
			WithClientName("Service"),
		),
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
	require.Equal(t, "app1", got.StaticClients[0].ID)
	require.Equal(t, []string{"http://a/cb", "http://b/cb"}, got.StaticClients[0].RedirectURIs)
	require.Equal(t, []string{"client_credentials"}, got.StaticClients[1].GrantTypes)
}

func TestRender_WithUsers_BcryptShape(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{mustUser(t, "u@e.com", "u", "p")}

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
	require.Equal(t, "u@e.com", p.Email)
	require.True(t, strings.HasPrefix(p.Hash, "$2a$") || strings.HasPrefix(p.Hash, "$2b$"), "bcrypt prefix")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(p.Hash), []byte("p")))
	require.NotEmpty(t, p.UserID, "userID should be auto-populated")
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
	require.Equal(t, "mockCallback", got.Connectors[0].Type)
}

func TestRender_NoAuthSource_Errors(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.enablePasswordDB = false
	// no connectors

	_, err := render(o)
	require.ErrorIs(t, err, ErrNoAuthSource)
}

func TestRender_IssuerRequired(t *testing.T) {
	o := defaultOptions()
	// issuer empty
	_, err := render(o)
	require.Error(t, err, "render must error when issuer is empty")
}

func TestRender_BcryptCost(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{mustUser(t, "u@e.com", "u", "p")}

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
	require.Equal(t, 10, cost, "bcrypt cost must be exactly 10: meets Dex minimum, fast enough for tests")
}

func TestRender_UserWithExplicitUserID(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	o.users = []User{mustUser(t, "u@e.com", "u", "p", WithUserID("fixed-id-123"))}

	out, err := render(o)
	require.NoError(t, err)

	var got struct {
		StaticPasswords []struct {
			UserID string `yaml:"userID"`
		} `yaml:"staticPasswords"`
	}
	require.NoError(t, yaml.Unmarshal(out, &got))
	require.Len(t, got.StaticPasswords, 1)
	require.Equal(t, "fixed-id-123", got.StaticPasswords[0].UserID)
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
	require.Equal(t, "local", got.OAuth2.PasswordConnector,
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
	require.Empty(t, got.OAuth2.PasswordConnector,
		"oauth2.passwordConnector must be omitted when enablePasswordDB is false")
}

func TestRender_YAMLInjection_NameField(t *testing.T) {
	o := defaultOptions()
	o.issuer = "http://h:5556/dex"
	// A name containing a newline + bogus YAML that would break a
	// template-based renderer. yaml.Marshal must escape it such that
	// the round-tripped value equals the input verbatim.
	malicious := "real-name\nmalicious_key: poison"
	o.clients = []Client{mustClient(t, "c", WithClientSecret("s"), WithClientName(malicious))}

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
	require.Equal(t, malicious, got.StaticClients[0].Name, "injected characters must round-trip as data, not structure")
	require.Empty(t, got.Malicious, "structural injection must not create a top-level key")
}

func TestDexLogLevel(t *testing.T) {
	cases := map[slog.Level]string{
		slog.LevelDebug:     "debug",
		slog.LevelDebug - 1: "debug",
		slog.LevelInfo:      "info",
		slog.LevelInfo + 1:  "warn",
		slog.LevelWarn:      "warn",
		slog.LevelWarn + 1:  "error",
		slog.LevelError:     "error",
	}
	for in, want := range cases {
		require.Equal(t, want, dexLogLevel(in), "slog.Level=%v", in)
	}
}
