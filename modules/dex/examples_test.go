package dex_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/guilycst/testcontainers-go/modules/dex"
	"golang.org/x/oauth2"
)

func ExampleRun_authorizationCode() {
	ctx := context.Background()
	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
		dex.WithClient(dex.Client{
			ID:           "my-app",
			Secret:       "secret",
			RedirectURIs: []string{"http://localhost:8080/callback"},
			GrantTypes:   []string{"authorization_code", "refresh_token"},
			Name:         "My App",
		}),
		dex.WithUser(dex.User{
			Email: "u@example.com", Username: "u", Password: "p",
		}),
	)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
	defer testcontainers.TerminateContainer(c)

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
	c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
		dex.WithClient(dex.Client{
			ID: "svc", Secret: "s", Name: "Service",
			GrantTypes: []string{"password"},
		}),
		dex.WithUser(dex.User{
			Email: "svc@svc.local", Username: "svc", Password: "svc-secret",
		}),
	)
	if err != nil {
		log.Fatalf("run: %v", err)
	}
	defer testcontainers.TerminateContainer(c)

	cfg := oauth2.Config{
		ClientID: "svc", ClientSecret: "s",
		Endpoint: oauth2.Endpoint{TokenURL: c.TokenEndpoint()},
		Scopes:   []string{"openid"},
	}
	tok, err := cfg.PasswordCredentialsToken(ctx, "svc@svc.local", "svc-secret")
	if err != nil {
		log.Fatalf("token: %v", err)
	}
	fmt.Println("has access token:", tok.AccessToken != "")
	// Output: has access token: true
}
