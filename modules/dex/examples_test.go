package dex_test

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/oauth2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
)

func ExampleRun() {
	ctx := context.Background()

	dexContainer, err := dex.Run(
		ctx,
		"dexidp/dex:v2.43.1-distroless",
		// explicitly declare the issuer URL - will be used in the OpenID Connect provider configuration
		// only necessary if you want to fetch the OpenID configuration directly from the Dex instance
		// alternatively there's a convenience method that monkey patches the discovery document
		// to match the mapped port
		dex.WithIssuer("http://localhost:15556"),
		// expose the corresponding port to match the issuer URL
		testcontainers.WithExposedPorts("15556:5556/tcp"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(dexContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// Register a client application
	clientApp, err := dexContainer.CreateClientApp(ctx, dex.WithClientName("testcontainers-go"))
	if err != nil {
		log.Printf("failed to create client app: %s", err)
		return
	}

	const (
		email    = "ted.tester@testcontainers.com"
		password = "$up3r$3crEt"
	)

	err = dexContainer.CreatePassword(
		ctx,
		dex.PlainTextCredential(email, password),
		dex.WithUsername("ted.tester"),
	)
	// Register an identity that can be used to authenticate with Dex
	if err != nil {
		log.Printf("failed to create identity: %s", err)
		return
	}

	oidcCfg, err := dexContainer.OpenIDConfiguration(ctx)
	if err != nil {
		log.Printf("failed to get OpenID configuration: %s", err)
		return
	}

	oauth2Cfg := oauth2.Config{
		ClientID:     clientApp.Id,
		ClientSecret: clientApp.Secret,
		Endpoint: oauth2.Endpoint{
			// Dex expects ClientID & Secret in the HTTP headers
			AuthStyle:     oauth2.AuthStyleInHeader,
			AuthURL:       oidcCfg.AuthorizationEndpoint,
			DeviceAuthURL: oidcCfg.DeviceAuthorizationEndpoint,
			TokenURL:      oidcCfg.TokenEndpoint,
		},
		// does not matter in this context
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      []string{"openid"},
	}

	// the primary identifier is always the email address **not** the username 🤷‍♂️
	tokenResp, err := oauth2Cfg.PasswordCredentialsToken(ctx, email, password)
	if err != nil {
		log.Printf("failed to get token response: %s", err)
		return
	}

	fmt.Println(tokenResp.TokenType)
	fmt.Println(tokenResp.Valid())

	// Output:
	// bearer
	// true
}
