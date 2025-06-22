package dex

// OpenIDConfiguration represents the OpenID Connect discovery document.
// Every OIDC provider must provide a valid OpenIDConfiguration.
// This struct contains all the necessary information for a client to interact with an OIDC provider - in this case, Dex.
type OpenIDConfiguration struct {
	// Issuer - name of the issuer, typically http://localhost:5556
	Issuer string `json:"issuer,omitzero"`
	// AuthorizationEndpoint - endpoint for authorization requests
	AuthorizationEndpoint string `json:"authorization_endpoint,omitzero"`
	// TokenEndpoint - endpoint for token requests (e.g. when using client credentials)
	TokenEndpoint string `json:"token_endpoint,omitzero"`
	// JWKSURI - endpoint for JSON Web Key Set (JWKS) requests
	JwksURI string `json:"jwks_uri,omitzero"`
	// UserInfoEndpoint - endpoint for user info requests
	UserinfoEndpoint string `json:"userinfo_endpoint,omitzero"`
	// DeviceAuthorizationEndpoint - endpoint for device authorization requests
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitzero"`
	// IntrospectionEndpoint - endpoint for token introspection requests
	IntrospectionEndpoint string `json:"introspection_endpoint,omitzero"`
	// GrantTypesSupported - list of grant types this provider supports
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`
	// ResponseTypesSupported - list of response types this provider supports
	ResponseTypesSupported []string `json:"response_types_supported,omitempty"`
	// SubjectTypesSupported - list of subject types this provider supports
	SubjectTypesSupported []string `json:"subject_types_supported,omitempty"`
	// IDTokenSigningAlgValues - list of signing algorithms this provider supports for ID tokens
	IDTokenSigningAlgValues []string `json:"id_token_signing_alg_values_supported,omitempty"`
	// CodeChallengeMethods - list of code challenge methods this provider supports
	CodeChallengeMethods []string `json:"code_challenge_methods_supported,omitempty"`
	// ScopesSupported - list of scopes this provider supports
	ScopesSupported []string `json:"scopes_supported,omitempty"`
	// TokenEndpointAuthMethods - list of token endpoint authentication methods this provider supports
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	// ClaimsSupported - list of claims this provider supports
	ClaimsSupported []string `json:"claims_supported,omitempty"`
}
