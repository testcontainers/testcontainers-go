package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func run() error {
	ctx := context.Background()

	connURL := os.Getenv("CONNECTION_URL")
	log.Printf("Using Lowkey Vault endpoint: %s", connURL)
	tokenURL := os.Getenv("IDENTITY_ENDPOINT")
	log.Printf("Using token URL: %s", tokenURL)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := http.Client{Transport: customTransport}

	resp, err := httpClient.Get(tokenURL + "?resource=" + connURL)
	if err != nil {
		log.Fatalf("failed to get token from token URL %v", err)
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Fatalf("failed to get token from token URL 2 %v", err)
		return err
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil) // Will use Managed Identity via the Assumed Identity container
	if err != nil {
		log.Fatalf("failed to create credential: %v", err)
		return err
	}
	secretClient, err := azsecrets.NewClient(connURL,
		cred,
		&azsecrets.ClientOptions{ClientOptions: policy.ClientOptions{Transport: &httpClient}, DisableChallengeResourceVerification: true})
	if err != nil {
		log.Fatalf("failed to create secret client: %v", err)
		return err
	}

	secretName := "secret-name"
	secretValue := "a secret value"
	created, err := secretClient.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{Value: &secretValue}, nil)
	if err != nil {
		log.Fatalf("failed to set the secret %v", err)
		return err
	}

	fetched, err := secretClient.GetSecret(ctx, secretName, created.ID.Version(), nil)
	if err != nil {
		log.Fatalf("failed to get the secret %v", err)
		return err
	}
	fetchedValue := *fetched.Value

	fmt.Println(fetchedValue == secretValue)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
