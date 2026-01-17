package lowkeyvalt_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/lowkeyvault"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"software.sslmate.com/src/go-pkcs12"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvalt.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	state, err := lowkeyVaultContainer.State(ctx)
	require.NoError(t, err)

	require.True(t, state.Running)
}

func TestRun_secretOperationsNetwork(t *testing.T) {
	ctx := context.Background()

	aNetwork, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, aNetwork)

	lowkeyVaultContainer, err := lowkeyvalt.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal",
		lowkeyvalt.WithNetworkAlias("lowkey-vault", aNetwork),
	)
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvalt.Network)
	require.NoError(t, err)
	require.NotNil(t, connUrl)

	tokenUrl, err := lowkeyVaultContainer.TokenUrl(ctx, lowkeyvalt.Network)
	require.NoError(t, err)
	require.NotNil(t, tokenUrl)

	networkContainer, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(
			testcontainers.FromDockerfile{
				Context:    "testdata",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
		testcontainers.WithEnv(map[string]string{
			"IDENTITY_ENDPOINT": tokenUrl,
			"IDENTITY_HEADER":   "header",
			"CONNECTION_URL":    connUrl,
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

	lowkeyVaultContainer, err := lowkeyvalt.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvalt.Local)
	require.NoError(t, err)
	require.NotNil(t, connUrl)

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	require.NoError(t, err)

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, _ := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	secretClient, _ := azsecrets.NewClient(connUrl,
		cred,
		&azsecrets.ClientOptions{ClientOptions: struct {
			APIVersion                      string
			Cloud                           cloud.Configuration
			InsecureAllowCredentialWithHTTP bool
			Logging                         policy.LogOptions
			Retry                           policy.RetryOptions
			Telemetry                       policy.TelemetryOptions
			TracingProvider                 tracing.Provider
			Transport                       policy.Transporter
			PerCallPolicies                 []policy.Policy
			PerRetryPolicies                []policy.Policy
		}{Transport: &httpClient}, DisableChallengeResourceVerification: true})

	secretName := "secret-name"
	secretValue := "a secret value"
	created, err := secretClient.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{Value: &secretValue}, nil)
	require.NoError(t, err)
	require.NotNil(t, created)

	fetched, err := secretClient.GetSecret(ctx, secretName, created.ID.Version(), nil)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	fetchedValue := *fetched.Secret.Value

	require.Equal(t, secretValue, fetchedValue)
}

func TestRun_keyOperations(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvalt.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvalt.Local)
	require.NoError(t, err)
	require.NotNil(t, connUrl)

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	require.NoError(t, err)

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, _ := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	keyClient, _ := azkeys.NewClient(connUrl,
		cred,
		&azkeys.ClientOptions{ClientOptions: struct {
			APIVersion                      string
			Cloud                           cloud.Configuration
			InsecureAllowCredentialWithHTTP bool
			Logging                         policy.LogOptions
			Retry                           policy.RetryOptions
			Telemetry                       policy.TelemetryOptions
			TracingProvider                 tracing.Provider
			Transport                       policy.Transporter
			PerCallPolicies                 []policy.Policy
			PerRetryPolicies                []policy.Policy
		}{Transport: &httpClient}, DisableChallengeResourceVerification: true})

	keyName := "rsa-key"
	rsaKeyParams := azkeys.CreateKeyParameters{
		Kty:     to.Ptr(azkeys.KeyTypeRSA),
		KeySize: to.Ptr(int32(2048)),
		KeyOps: []*azkeys.KeyOperation{
			to.Ptr(azkeys.KeyOperationDecrypt),
			to.Ptr(azkeys.KeyOperationEncrypt),
			to.Ptr(azkeys.KeyOperationUnwrapKey),
			to.Ptr(azkeys.KeyOperationWrapKey),
		},
	}
	createdKey, err := keyClient.CreateKey(ctx, keyName, rsaKeyParams, nil)
	require.NoError(t, err)
	require.NotNil(t, createdKey)

	secretMessage := "a secret message"
	encryptionParameters := azkeys.KeyOperationParameters{
		Value:     []byte(secretMessage),
		Algorithm: to.Ptr(azkeys.EncryptionAlgorithmRSAOAEP256)}
	encrResp, err := keyClient.Encrypt(ctx, keyName, createdKey.Key.KID.Version(), encryptionParameters, nil)
	require.NoError(t, err)
	require.NotNil(t, encrResp)
	cipherText := encrResp.Result

	decryptionParameters := azkeys.KeyOperationParameters{
		Value:     cipherText,
		Algorithm: to.Ptr(azkeys.EncryptionAlgorithmRSAOAEP256)}
	decrResp, err := keyClient.Decrypt(ctx, keyName, createdKey.Key.KID.Version(), decryptionParameters, nil)
	require.NoError(t, err)
	require.NotNil(t, decrResp)
	decryptedMessage := string(decrResp.Result)

	require.Equal(t, secretMessage, decryptedMessage)
}

func TestRun_certificateOperations(t *testing.T) {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvalt.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	testcontainers.CleanupContainer(t, lowkeyVaultContainer)
	require.NoError(t, err)

	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvalt.Local)
	require.NoError(t, err)
	require.NotNil(t, connUrl)

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	require.NoError(t, err)

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, _ := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	certClient, _ := azcertificates.NewClient(connUrl,
		cred,
		&azcertificates.ClientOptions{ClientOptions: struct {
			APIVersion                      string
			Cloud                           cloud.Configuration
			InsecureAllowCredentialWithHTTP bool
			Logging                         policy.LogOptions
			Retry                           policy.RetryOptions
			Telemetry                       policy.TelemetryOptions
			TracingProvider                 tracing.Provider
			Transport                       policy.Transporter
			PerCallPolicies                 []policy.Policy
			PerRetryPolicies                []policy.Policy
		}{Transport: &httpClient}, DisableChallengeResourceVerification: true})

	certName := "ec-cert"
	subject := "CN=example.com"
	created, err := certClient.CreateCertificate(ctx, certName, azcertificates.CreateCertificateParameters{
		CertificatePolicy: &azcertificates.CertificatePolicy{
			IssuerParameters: &azcertificates.IssuerParameters{
				Name: to.Ptr("Self"),
			},
			KeyProperties: &azcertificates.KeyProperties{
				Curve:    to.Ptr(azcertificates.CurveNameP256),
				KeyType:  to.Ptr(azcertificates.KeyTypeEC),
				ReuseKey: to.Ptr(true),
			},
			SecretProperties: &azcertificates.SecretProperties{
				ContentType: to.Ptr("application/x-pkcs12"),
			},
			X509CertificateProperties: &azcertificates.X509CertificateProperties{
				Subject: &subject,
				SubjectAlternativeNames: &azcertificates.SubjectAlternativeNames{
					DNSNames: []*string{to.Ptr("localhost")},
				},
				ValidityInMonths: to.Ptr(int32(12)),
			},
		},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, created)

	secretClient, _ := azsecrets.NewClient(connUrl,
		cred,
		&azsecrets.ClientOptions{ClientOptions: struct {
			APIVersion                      string
			Cloud                           cloud.Configuration
			InsecureAllowCredentialWithHTTP bool
			Logging                         policy.LogOptions
			Retry                           policy.RetryOptions
			Telemetry                       policy.TelemetryOptions
			TracingProvider                 tracing.Provider
			Transport                       policy.Transporter
			PerCallPolicies                 []policy.Policy
			PerRetryPolicies                []policy.Policy
		}{Transport: &httpClient}, DisableChallengeResourceVerification: true})

	base64Secret, err := secretClient.GetSecret(ctx, certName, "", nil)
	require.NoError(t, err)
	require.NotNil(t, base64Secret)
	base64Value := *base64Secret.Secret.Value
	bytes, err := base64.StdEncoding.DecodeString(base64Value)
	require.NoError(t, err)
	// use SSLMate library to decode the certificate store as the x/crypto
	// library is not fully compatible with the Java PKCS12 format
	key, cert, err := pkcs12.Decode(bytes, "")
	require.NoError(t, err)
	ecKey := key.(*ecdsa.PrivateKey)

	require.Equal(t, subject, cert.Subject.String())
	require.Equal(t, elliptic.P256(), ecKey.Curve)
}
