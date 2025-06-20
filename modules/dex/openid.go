package dex

type OpenIDConfiguration struct {
	Issuer                      string   `json:"issuer,omitzero"`
	AuthorizationEndpoint       string   `json:"authorization_endpoint,omitzero"`
	TokenEndpoint               string   `json:"token_endpoint,omitzero"`
	JwksURI                     string   `json:"jwks_uri,omitzero"`
	UserinfoEndpoint            string   `json:"userinfo_endpoint,omitzero"`
	DeviceAuthorizationEndpoint string   `json:"device_authorization_endpoint,omitzero"`
	IntrospectionEndpoint       string   `json:"introspection_endpoint,omitzero"`
	GrantTypesSupported         []string `json:"grant_types_supported,omitempty"`
	ResponseTypesSupported      []string `json:"response_types_supported,omitempty"`
	SubjectTypesSupported       []string `json:"subject_types_supported,omitempty"`
	IDTokenSigningAlgValues     []string `json:"id_token_signing_alg_values_supported,omitempty"`
	CodeChallengeMethods        []string `json:"code_challenge_methods_supported,omitempty"`
	ScopesSupported             []string `json:"scopes_supported,omitempty"`
	TokenEndpointAuthMethods    []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ClaimsSupported             []string `json:"claims_supported,omitempty"`
}
