package dex

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Defaults(t *testing.T) {
	o := defaultOptions()
	assert.True(t, o.skipApprovalScreen)
	assert.Equal(t, StorageSQLite, o.storage)
	assert.Equal(t, slog.LevelInfo, o.logLevel)
	assert.True(t, o.enablePasswordDB)
	assert.Empty(t, o.issuer)
}

func TestOptions_Apply(t *testing.T) {
	o := defaultOptions()

	clientA, err := NewClient("a", WithClientSecret("s"))
	require.NoError(t, err)
	clientB, err := NewClient("b")
	require.NoError(t, err)
	user, err := NewUser("u@e.com", "u", "p")
	require.NoError(t, err)

	applyAll := []Option{
		WithClient(clientA),
		WithClient(clientB),
		WithUser(user),
		WithConnector(ConnectorMock, "m", "Mock"),
		WithIssuer("http://dex:5556/dex"),
		WithSkipApprovalScreen(false),
		WithStorage(StorageMemory),
		WithLogLevel(slog.LevelDebug),
		WithLogger(slog.Default()),
		WithDisablePasswordDB(),
	}
	for _, opt := range applyAll {
		require.NoError(t, opt(&o))
	}

	assert.Len(t, o.clients, 2)
	assert.Equal(t, "b", o.clients[1].id)
	assert.Len(t, o.users, 1)
	assert.Len(t, o.connectors, 1)
	assert.Equal(t, ConnectorMock, o.connectors[0].Type)
	assert.Equal(t, "m", o.connectors[0].ID)
	assert.Equal(t, "Mock", o.connectors[0].Name)
	assert.Equal(t, "http://dex:5556/dex", o.issuer)
	assert.False(t, o.skipApprovalScreen)
	assert.Equal(t, StorageMemory, o.storage)
	assert.Equal(t, slog.LevelDebug, o.logLevel)
	assert.NotNil(t, o.logger)
	assert.False(t, o.enablePasswordDB)
}

func TestOptions_WithConnectorPassword_IsNoOp(t *testing.T) {
	o := defaultOptions()
	require.NoError(t, WithConnector(ConnectorPassword, "local", "Local")(&o))
	assert.Empty(t, o.connectors, "ConnectorPassword must not be appended; password DB covers it")
	assert.True(t, o.enablePasswordDB, "password DB stays enabled by default")
}

func TestOptions_WithConnector_RejectsBlankFields(t *testing.T) {
	o := defaultOptions()

	err := WithConnector(ConnectorMock, "", "Mock")(&o)
	assert.Error(t, err, "blank id must be rejected")

	err = WithConnector(ConnectorMock, "id", "")(&o)
	assert.Error(t, err, "blank name must be rejected")

	assert.Empty(t, o.connectors, "rejected options must not mutate state")
}

func TestOptions_WithIssuer_RejectsBlank(t *testing.T) {
	o := defaultOptions()
	assert.Error(t, WithIssuer("")(&o))
	assert.Empty(t, o.issuer)
}

func TestOptions_WithStorage_RejectsBlank(t *testing.T) {
	o := defaultOptions()
	assert.Error(t, WithStorage("")(&o))
	assert.Equal(t, StorageSQLite, o.storage, "failed option must leave default intact")
}

func TestOptions_WithClient_RejectsZeroValue(t *testing.T) {
	o := defaultOptions()
	assert.Error(t, WithClient(Client{})(&o), "zero Client must be rejected — force NewClient")
	assert.Empty(t, o.clients)
}

func TestOptions_WithUser_RejectsZeroValue(t *testing.T) {
	o := defaultOptions()
	assert.Error(t, WithUser(User{})(&o), "zero User must be rejected — force NewUser")
	assert.Empty(t, o.users)
}

func TestNewClient_WithClientGrantTypes_Allowlist(t *testing.T) {
	cases := []struct {
		name    string
		grants  []string
		wantErr bool
	}{
		{"all four supported", []string{"authorization_code", "refresh_token", "client_credentials", "password"}, false},
		{"single supported", []string{"authorization_code"}, false},
		{"blank rejected", []string{""}, true},
		{"typo rejected", []string{"authorisation_code"}, true},
		{"unknown grant rejected", []string{"jwt-bearer"}, true},
		{"valid then invalid rejected", []string{"password", "jwt-bearer"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient("c", WithClientSecret("s"), WithClientGrantTypes(tc.grants...))
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
