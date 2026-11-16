package dex_test

import (
	"context"
	"fmt"
	"log"

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

	fmt.Println("has issuer:", c.IssuerURL() != "")
	// Output: has issuer: true
}
