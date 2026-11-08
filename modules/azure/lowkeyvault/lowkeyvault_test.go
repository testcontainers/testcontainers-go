package lowkeyvault_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/lowkeyvault"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	state, err := lowkeyVaultContainer.State(ctx)
	require.NoError(t, err)

	require.True(t, state.Running)
}

func TestRun_secretOperationsNetwork(t *testing.T) {
	ctx := context.Background()

	aNetwork, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, aNetwork)
	require.NoError(t, err)

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal",
		lowkeyvault.WithNetworkAlias("lowkey-vault", aNetwork),
	)
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connURL, err := lowkeyVaultContainer.ConnectionURL(ctx, lowkeyvault.Network)
	require.NoError(t, err)
	require.NotNil(t, connURL)

	tokenURL, err := lowkeyVaultContainer.IdentityEndpoint(ctx, lowkeyvault.Network)
	require.NoError(t, err)
	require.NotNil(t, tokenURL)

	networkContainer, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(
			testcontainers.FromDockerfile{
				Context:    "testdata",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
		testcontainers.WithEnv(map[string]string{
			"IDENTITY_ENDPOINT": tokenURL,
			"IDENTITY_HEADER":   lowkeyVaultContainer.IdentityHeader(),
			"CONNECTION_URL":    connURL,
		}),
		network.WithNetwork(nil, aNetwork),
		testcontainers.WithWaitStrategy(
			wait.ForLog("true"),
		),
	)
	testcontainers.CleanupContainer(t, networkContainer)
	require.NoError(t, err)
	require.NotNil(t, networkContainer)
	state, err := networkContainer.State(ctx)
	require.NoError(t, err)

	require.Equal(t, 0, state.ExitCode)
}

func TestRun_secretOperations(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connURL, err := lowkeyVaultContainer.ConnectionURL(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	require.NotNil(t, connURL)

	identityEndpoint, err := lowkeyVaultContainer.IdentityEndpoint(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	t.Setenv("IDENTITY_ENDPOINT", identityEndpoint)
	t.Setenv("IDENTITY_HEADER", lowkeyVaultContainer.IdentityHeader())

	httpClient, err := lowkeyVaultContainer.Client(ctx)
	require.NoError(t, err)

	httpClient.Timeout = 30 * time.Second
	var token string

	secretName := "secret-name"
	secretValue := "a secret value"
	var created struct {
		ID    string  `json:"id"`
		Value *string `json:"value"`
	}
	response := keyVaultRequest(t, ctx, &httpClient, &token, http.MethodPut,
		connURL+"/secrets/"+url.PathEscape(secretName)+"?api-version=2025-07-01",
		map[string]string{"value": secretValue}, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &created))
	require.NotEmpty(t, created.ID)
	createdURL, err := url.Parse(created.ID)
	require.NoError(t, err)
	idParts := strings.Split(strings.Trim(createdURL.Path, "/"), "/")
	require.Len(t, idParts, 3)
	require.NotEmpty(t, idParts[2])

	var fetched struct {
		Value *string `json:"value"`
	}
	response = keyVaultRequest(t, ctx, &httpClient, &token, http.MethodGet,
		connURL+"/secrets/"+url.PathEscape(secretName)+"/"+url.PathEscape(idParts[2])+"?api-version=2025-07-01",
		nil, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &fetched))
	require.NotNil(t, fetched.Value)
	fetchedValue := *fetched.Value
	require.Equal(t, secretValue, fetchedValue)
}

func TestRun_keyOperations(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connURL, err := lowkeyVaultContainer.ConnectionURL(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	require.NotNil(t, connURL)

	identityEndpoint, err := lowkeyVaultContainer.IdentityEndpoint(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	t.Setenv("IDENTITY_ENDPOINT", identityEndpoint)
	t.Setenv("IDENTITY_HEADER", lowkeyVaultContainer.IdentityHeader())

	httpClient, err := lowkeyVaultContainer.Client(ctx)
	require.NoError(t, err)

	httpClient.Timeout = 30 * time.Second
	var token string

	keyName := "rsa-key"
	rsaKeyParams := map[string]any{
		"kty":      "RSA",
		"key_size": 2048,
		"key_ops":  []string{"decrypt", "encrypt", "unwrapKey", "wrapKey"},
	}
	var createdKey struct {
		Key *struct {
			KID string `json:"kid"`
		} `json:"key"`
	}
	response := keyVaultRequest(t, ctx, &httpClient, &token, http.MethodPost,
		connURL+"/keys/"+url.PathEscape(keyName)+"/create?api-version=7.6", rsaKeyParams, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &createdKey))
	require.NotNil(t, createdKey.Key)
	require.NotEmpty(t, createdKey.Key.KID)
	keyURL, err := url.Parse(createdKey.Key.KID)
	require.NoError(t, err)
	idParts := strings.Split(strings.Trim(keyURL.Path, "/"), "/")
	require.Len(t, idParts, 3)
	require.NotEmpty(t, idParts[2])
	keyEndpoint := connURL + "/keys/" + url.PathEscape(keyName) + "/" + url.PathEscape(idParts[2])

	secretMessage := "a secret message"
	encryptionParameters := map[string]string{
		"value": base64.RawURLEncoding.EncodeToString([]byte(secretMessage)),
		"alg":   "RSA-OAEP-256",
	}
	var encrResp struct {
		Value string `json:"value"`
	}
	response = keyVaultRequest(t, ctx, &httpClient, &token, http.MethodPost,
		keyEndpoint+"/encrypt?api-version=7.6", encryptionParameters, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &encrResp))
	cipherText, err := base64.RawURLEncoding.DecodeString(encrResp.Value)
	require.NoError(t, err)
	require.NotEmpty(t, cipherText)

	decryptionParameters := map[string]string{
		"value": base64.RawURLEncoding.EncodeToString(cipherText),
		"alg":   "RSA-OAEP-256",
	}
	var decrResp struct {
		Value string `json:"value"`
	}
	response = keyVaultRequest(t, ctx, &httpClient, &token, http.MethodPost,
		keyEndpoint+"/decrypt?api-version=7.6", decryptionParameters, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &decrResp))
	plainText, err := base64.RawURLEncoding.DecodeString(decrResp.Value)
	require.NoError(t, err)
	decryptedMessage := string(plainText)
	require.Equal(t, secretMessage, decryptedMessage)
}

func TestRun_certificateOperations(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.3.48-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connURL, err := lowkeyVaultContainer.ConnectionURL(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	require.NotNil(t, connURL)

	identityEndpoint, err := lowkeyVaultContainer.IdentityEndpoint(ctx, lowkeyvault.Local)
	require.NoError(t, err)
	t.Setenv("IDENTITY_ENDPOINT", identityEndpoint)
	t.Setenv("IDENTITY_HEADER", lowkeyVaultContainer.IdentityHeader())

	httpClient, err := lowkeyVaultContainer.Client(ctx)
	require.NoError(t, err)

	httpClient.Timeout = 30 * time.Second
	var token string

	certName := "ec-cert"
	subject := "CN=example.com"
	certificateParameters := map[string]any{
		"policy": map[string]any{
			"issuer":       map[string]string{"name": "Self"},
			"key_props":    map[string]any{"crv": "P-256", "kty": "EC", "reuse_key": true},
			"secret_props": map[string]string{"contentType": "application/x-pkcs12"},
			"x509_props": map[string]any{
				"subject":         subject,
				"sans":            map[string][]string{"dns_names": {"localhost"}},
				"validity_months": 12,
			},
		},
	}
	response := keyVaultRequest(t, ctx, &httpClient, &token, http.MethodPost,
		connURL+"/certificates/"+url.PathEscape(certName)+"/create?api-version=7.6",
		certificateParameters, http.StatusAccepted)
	var created map[string]any
	require.NoError(t, json.Unmarshal(response, &created))
	require.NotNil(t, created)

	var base64Secret struct {
		Value *string `json:"value"`
	}
	response = keyVaultRequest(t, ctx, &httpClient, &token, http.MethodGet,
		connURL+"/secrets/"+url.PathEscape(certName)+"/?api-version=2025-07-01", nil, http.StatusOK)
	require.NoError(t, json.Unmarshal(response, &base64Secret))
	require.NotNil(t, base64Secret.Value)
	base64Value := *base64Secret.Value
	bytes, err := base64.StdEncoding.DecodeString(base64Value)
	require.NoError(t, err)
	// use SSLMate to decode the certificate store because x/crypto
	// is not fully compatible with the Java PKCS12 format.
	key, cert, err := pkcs12.Decode(bytes, "")
	require.NoError(t, err)
	ecKey := key.(*ecdsa.PrivateKey)

	require.Equal(t, subject, cert.Subject.String())
	require.Equal(t, elliptic.P256(), ecKey.Curve)
}

// keyVaultRequest shares only HTTP transport and managed identity authentication;
// each TestRun function supplies its operations, API versions, and assertions.
func keyVaultRequest(t *testing.T, ctx context.Context, client *http.Client, token *string, method, endpoint string, payload any, status int) []byte {
	t.Helper()
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if *token == "" {
		// Like the SDK challenge policy, elicit the challenge without sending the body.
		probe := request.Clone(ctx)
		probe.Body = nil
		probe.GetBody = nil
		probe.ContentLength = 0
		probe.Header.Del("Content-Type")
		challenge, err := client.Do(probe)
		require.NoError(t, err)
		readKeyVaultResponse(t, challenge, http.StatusUnauthorized)

		// The SDK prefers scope over resource and removes /.default for managed identity.
		authHeader := challenge.Header.Get("WWW-Authenticate")
		require.True(t, strings.HasPrefix(authHeader, "Bearer "), "missing Bearer challenge")
		resource := ""
		for _, field := range []string{"scope", "resource"} {
			match := regexp.MustCompile(`(?:^|[ ,])` + field + `="([^"]+)"`).FindStringSubmatch(authHeader)
			if len(match) == 2 {
				resource = strings.TrimSuffix(match[1], "/.default")
				break
			}
		}
		require.NotEmpty(t, resource, "missing challenge resource")
		// The existing emulator SDK options disable challenge resource verification.
		identityURL, err := url.Parse(os.Getenv("IDENTITY_ENDPOINT"))
		require.NoError(t, err)
		query := identityURL.Query()
		query.Set("api-version", "2019-08-01")
		query.Set("resource", resource)
		identityURL.RawQuery = query.Encode()
		identityRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, identityURL.String(), nil)
		require.NoError(t, err)
		identityRequest.Header.Set("X-IDENTITY-HEADER", os.Getenv("IDENTITY_HEADER"))
		identityResponse, err := client.Do(identityRequest)
		require.NoError(t, err)
		var credential struct {
			AccessToken string `json:"access_token"`
		}
		require.NoError(t, json.Unmarshal(readKeyVaultResponse(t, identityResponse, http.StatusOK), &credential))
		require.NotEmpty(t, credential.AccessToken)
		*token = credential.AccessToken
	}
	request.Header.Set("Authorization", "Bearer "+*token)
	response, err := client.Do(request)
	require.NoError(t, err)
	return readKeyVaultResponse(t, response, status)
}

func readKeyVaultResponse(t *testing.T, response *http.Response, status int) []byte {
	t.Helper()
	defer response.Body.Close()
	const maxBody = 1 << 20
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(body), maxBody, "Key Vault response exceeds limit")
	require.Equal(t, status, response.StatusCode, "%s", body)
	return body
}
