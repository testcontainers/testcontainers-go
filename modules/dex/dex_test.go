package dex_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
)

func TestDex(t *testing.T) {
	ctx := context.Background()

	ctr, err := dex.Run(
		ctx,
		"dexidp/dex:v2.43.1-distroless",
		dex.WithIssuer("http://localhost:15556"),
		testcontainers.WithExposedPorts("15556:5556/tcp"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestDex_OpenIDConfiguration(t *testing.T) {
	ctx := context.Background()

	ctr, err := dex.Run(ctx, "dexidp/dex:v2.43.1-distroless")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	rawOIDCCfg, err := ctr.RawOpenIDConfiguration(ctx)
	require.NoError(t, err)

	var decodedOIDCCfg dex.OpenIDConfiguration
	err = json.Unmarshal(rawOIDCCfg, &decodedOIDCCfg)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:5556", decodedOIDCCfg.Issuer)
	assert.False(t, strings.HasPrefix(decodedOIDCCfg.AuthorizationEndpoint, "http://localhost:5556"))
}
