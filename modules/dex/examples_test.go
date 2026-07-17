package dex_test

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/oauth2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dex"
)

func ExampleRun_authorizationCode() {
	// runContainer {
	ctx := context.Background()

	app, err := dex.NewClient("my-app",
		dex.WithClientSecret("secret"),
		dex.WithClientRedirectURIs("http://localhost:8080/callback"),
		dex.WithClientGrantTypes("authorization_code", "refresh_token"),
		dex.WithClientName("My App"),
	)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}
	user, err := dex.NewUser("u@example.com", "u", "p")
	if err != nil {
		log.Fatalf("new user: %v", err)
	}

	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
		dex.WithClient(app),
		dex.WithUser(user),
	)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
	defer func() { _ = testcontainers.TerminateContainer(c) }()
	// }

	_ = oauth2.Config{
		ClientID:     "my-app",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     oauth2.Endpoint{AuthURL: c.AuthEndpoint(), TokenURL: c.TokenEndpoint()},
		Scopes:       []string{"openid", "email"},
	}
	fmt.Println("has issuer:", c.IssuerURL() != "")
	// Output: has issuer: true
}

func ExampleRun_passwordGrant() {
	// Dex's recommended machine-to-machine pattern: ROPC with a dedicated
	// service-account user. (client_credentials requires an upstream
	// connector — see module README.)
	ctx := context.Background()

	svc, err := dex.NewClient("svc",
		dex.WithClientSecret("s"),
		dex.WithClientName("Service"),
		dex.WithClientGrantTypes("password"),
	)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}
	user, err := dex.NewUser("svc@svc.local", "svc", "svc-secret")
	if err != nil {
		log.Fatalf("new user: %v", err)
	}

	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
		dex.WithClient(svc),
		dex.WithUser(user),
	)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
	defer func() { _ = testcontainers.TerminateContainer(c) }()

	cfg := oauth2.Config{
		ClientID: "svc", ClientSecret: "s",
		Endpoint: oauth2.Endpoint{TokenURL: c.TokenEndpoint()},
		Scopes:   []string{"openid"},
	}
	tok, err := cfg.PasswordCredentialsToken(ctx, "svc@svc.local", "svc-secret")
	if err != nil {
		panic(fmt.Errorf("token: %w", err))
	}
	fmt.Println("has access token:", tok.AccessToken != "")
	// Output: has access token: true
}
