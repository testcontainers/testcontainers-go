package lowkeyvault_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/lowkeyvault"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"software.sslmate.com/src/go-pkcs12"
)

func ExampleRun() {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}

	state, err := lowkeyVaultContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_secretOperationsNetwork() {
	ctx := context.Background()

	aNetwork, err := network.New(ctx)
	defer func() {
		err := aNetwork.Remove(ctx)
		if err != nil {
			log.Fatalf("failed to remove network: %s", err)
			return
		}
	}()
	if err != nil {
		log.Fatalf("failed to setup network: %s", err)
		return
	}

	// createContainerWithNetwork {
	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal",
		lowkeyvault.WithNetworkAlias("lowkey-vault", aNetwork),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}
	// }

	// obtainEndpointUrls {
	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvault.Network)
	if err != nil {
		log.Fatalf("failed to get connection url: %s", err)
		return
	}

	tokenUrl, err := lowkeyVaultContainer.TokenUrl(ctx, lowkeyvault.Network)
	if err != nil {
		log.Fatalf("failed to get token url: %s", err)
		return
	}
	// }

	networkContainer, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(
			testcontainers.FromDockerfile{
				Context:    "testdata",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
		// configureClient {
		testcontainers.WithEnv(map[string]string{
			"IDENTITY_ENDPOINT": tokenUrl,
			"IDENTITY_HEADER":   "header",
			"CONNECTION_URL":    connUrl,
		}),
		// }
		network.WithNetwork(nil, aNetwork),
		testcontainers.WithWaitStrategy(
			wait.ForLog("true"),
		),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(networkContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}
	state, err := networkContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.ExitCode == 0)

	// Output:
	// true
}

func ExampleRun_secretOperations() {
	ctx := context.Background()

	// createContainer {
	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}
	// }

	// prepareTheSecretClient {
	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvault.Local)
	if err != nil {
		log.Fatalf("failed to get connection url: %s", err)
		return
	}

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	if err != nil {
		log.Fatalf("failed to set managed identity variables: %s", err)
		return
	}

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, err := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	if err != nil {
		log.Fatalf("failed to create credential: %v", err)
		return
	}
	secretClient, err := azsecrets.NewClient(connUrl,
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
	if err != nil {
		log.Fatalf("failed to create secret client: %v", err)
		return
	}
	// }

	// setAndFetchTheSecret {
	secretName := "secret-name"
	secretValue := "a secret value"
	created, err := secretClient.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{Value: &secretValue}, nil)
	if err != nil {
		log.Fatalf("failed to set the secret %s", err.Error())
	}

	fetched, err := secretClient.GetSecret(ctx, secretName, created.ID.Version(), nil)
	if err != nil {
		log.Fatalf("failed to get the secret %s", err.Error())
	}
	fetchedValue := *fetched.Secret.Value
	// }

	fmt.Println(fetchedValue == secretValue)

	// Output:
	// true
}

func ExampleRun_keyOperations() {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}

	// prepareTheKeyClient {
	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvault.Local)
	if err != nil {
		log.Fatalf("failed to get connection url: %s", err)
		return
	}

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	if err != nil {
		log.Fatalf("failed to set managed identity variables: %s", err)
		return
	}

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, err := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	if err != nil {
		log.Fatalf("failed to create credential: %v", err)
		return
	}
	keyClient, err := azkeys.NewClient(connUrl,
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
	if err != nil {
		log.Fatalf("failed to create key client: %v", err)
		return
	}
	// }

	// createKey {
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
	if err != nil {
		log.Fatalf("failed to create a key: %v", err)
	}
	// }

	// encryptMessage {
	secretMessage := "a secret message"
	encryptionParameters := azkeys.KeyOperationParameters{
		Value:     []byte(secretMessage),
		Algorithm: to.Ptr(azkeys.EncryptionAlgorithmRSAOAEP256)}
	encrResp, err := keyClient.Encrypt(ctx, keyName, createdKey.Key.KID.Version(), encryptionParameters, nil)
	if err != nil {
		log.Fatalf("failed to encrypt a message: %v", err)
	}
	cipherText := encrResp.Result
	// }

	// decryptCipherText {
	decryptionParameters := azkeys.KeyOperationParameters{
		Value:     cipherText,
		Algorithm: to.Ptr(azkeys.EncryptionAlgorithmRSAOAEP256)}
	decrResp, err := keyClient.Decrypt(ctx, keyName, createdKey.Key.KID.Version(), decryptionParameters, nil)
	if err != nil {
		log.Fatalf("failed to encrypt a message: %v", err)
	}
	decryptedMessage := string(decrResp.Result)
	// }

	fmt.Println(decryptedMessage == secretMessage)

	// Output:
	// true
}

func ExampleRun_certificateOperations() {
	ctx := context.Background()

	lowkeyVaultContainer, err := lowkeyvault.Run(ctx, "nagyesta/lowkey-vault:7.0.9-ubi10-minimal")
	defer func() {
		if err := testcontainers.TerminateContainer(lowkeyVaultContainer); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}

	// prepareTheCertClient {
	connUrl, err := lowkeyVaultContainer.ConnectionUrl(ctx, lowkeyvault.Local)
	if err != nil {
		log.Fatalf("failed to get connection url: %s", err)
		return
	}

	err = lowkeyVaultContainer.SetManagedIdentityEnvVariables(ctx)
	if err != nil {
		log.Fatalf("failed to set managed identity variables: %s", err)
		return
	}

	httpClient := lowkeyVaultContainer.PrepareClientForSelfSignedCert()

	cred, err := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	if err != nil {
		log.Fatalf("failed to create credential: %v", err)
		return
	}
	certClient, err := azcertificates.NewClient(connUrl,
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
	if err != nil {
		log.Fatalf("failed to create certificate client: %v", err)
		return
	}
	// }

	// createCertificate {
	certName := "ec-cert"
	subject := "CN=example.com"
	_, err = certClient.CreateCertificate(ctx, certName, azcertificates.CreateCertificateParameters{
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
	if err != nil {
		log.Fatalf("failed to create a certificate: %v", err)
	}
	// }

	secretClient, err := azsecrets.NewClient(connUrl,
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
	if err != nil {
		log.Fatalf("failed to create secret client: %v", err)
		return
	}

	// fetchCertDetails {
	base64Secret, err := secretClient.GetSecret(ctx, certName, "", nil)
	if err != nil {
		log.Fatalf("failed to get the secret with certificate store: %v", err)
		return
	}
	base64Value := *base64Secret.Secret.Value
	bytes, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		log.Fatalf("failed to decode the certificate store: %v", err)
		return
	}
	// use SSLMate library to decode the certificate store as the x/crypto
	// library is not fully compatible with the Java PKCS12 format
	key, cert, err := pkcs12.Decode(bytes, "")
	if err != nil {
		log.Fatalf("failed to open certificate store: %v", err)
		return
	}
	ecKey := key.(*ecdsa.PrivateKey)
	// }

	fmt.Println(cert.Subject.String() == subject && ecKey.Curve == elliptic.P256())

	// Output:
	// true
}
