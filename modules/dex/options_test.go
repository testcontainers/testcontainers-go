package dex

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Defaults(t *testing.T) {
	o := defaultOptions()
	assert.True(t, o.skipApprovalScreen)
	assert.Equal(t, "sqlite3", o.storage)
	assert.Equal(t, "info", o.logLevel)
	assert.True(t, o.enablePasswordDB)
	assert.Empty(t, o.issuer)
}

func TestOptions_Apply(t *testing.T) {
	o := defaultOptions()
	applyAll := []Option{
		WithClient(Client{ID: "a", Secret: "s"}),
		WithClient(Client{ID: "b"}),
		WithUser(User{Email: "u@e.com", Username: "u", Password: "p"}),
		WithConnector(ConnectorMock, "m", "Mock"),
		WithIssuer("http://dex:5556/dex"),
		WithSkipApprovalScreen(false),
		WithStorage("memory"),
		WithLogLevel("debug"),
		WithLogger(slog.Default()),
	}
	for _, opt := range applyAll {
		opt(&o)
	}

	assert.Len(t, o.clients, 2)
	assert.Equal(t, "b", o.clients[1].ID)
	assert.Len(t, o.users, 1)
	assert.Len(t, o.connectors, 1)
	assert.Equal(t, ConnectorMock, o.connectors[0].Type)
	assert.Equal(t, "m", o.connectors[0].ID)
	assert.Equal(t, "Mock", o.connectors[0].Name)
	assert.Equal(t, "http://dex:5556/dex", o.issuer)
	assert.False(t, o.skipApprovalScreen)
	assert.Equal(t, "memory", o.storage)
	assert.Equal(t, "debug", o.logLevel)
	assert.NotNil(t, o.logger)
}

func TestOptions_WithConnectorPassword_IsNoOp(t *testing.T) {
	o := defaultOptions()
	WithConnector(ConnectorPassword, "local", "Local")(&o)
	assert.Empty(t, o.connectors, "ConnectorPassword must not be appended; password DB covers it")
	assert.True(t, o.enablePasswordDB, "password DB stays enabled by default")
}
