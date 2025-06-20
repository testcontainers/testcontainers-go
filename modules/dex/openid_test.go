package dex_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/dex"
)

var (
	//go:embed testdata/openid_configuration.json
	jsonConfig []byte

	dexOpenIDConfig = dex.OpenIDConfiguration{
		Issuer:                      "http://localhost:15556",
		AuthorizationEndpoint:       "http://localhost:15556/auth",
		TokenEndpoint:               "http://localhost:15556/token",
		JwksURI:                     "http://localhost:15556/keys",
		UserinfoEndpoint:            "http://localhost:15556/userinfo",
		DeviceAuthorizationEndpoint: "http://localhost:15556/device/code",
		IntrospectionEndpoint:       "http://localhost:15556/token/introspect",
		GrantTypesSupported:         []string{"authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code", "urn:ietf:params:oauth:grant-type:token-exchange"},
		ResponseTypesSupported:      []string{"code"},
		SubjectTypesSupported:       []string{"public"},
		IDTokenSigningAlgValues:     []string{"RS256"},
		CodeChallengeMethods:        []string{"S256", "plain"},
		ScopesSupported:             []string{"openid", "email", "groups", "profile", "offline_access"},
		TokenEndpointAuthMethods:    []string{"client_secret_basic", "client_secret_post"},
		ClaimsSupported:             []string{"iss", "sub", "aud", "iat", "exp", "email", "email_verified", "locale", "name", "preferred_username", "at_hash"},
	}
)

func TestOpenIDConfiguration_JsonUnmarshal(t *testing.T) {
	var unmarshaledConfig dex.OpenIDConfiguration

	require.NoError(t, json.Unmarshal(jsonConfig, &unmarshaledConfig))
	require.Equal(t, dexOpenIDConfig, unmarshaledConfig)
}
