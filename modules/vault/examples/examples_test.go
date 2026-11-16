package examples_test

import (
	"context"
	"fmt"
	"log"
	"time"

	vaultClient "github.com/hashicorp/vault-client-go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func ExampleRun_clientRead() {
	// clientLibRead {
	ctx := context.Background()
	const token = "root-token"

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.0",
		vault.WithToken(token),
		vault.WithInitCommand("secrets enable transit", "write -f transit/keys/my-key"),
		vault.WithInitCommand("kv put secret/test1 foo1=bar1"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	hostAddress, err := vaultContainer.HttpHostAddress(ctx)
	if err != nil {
		log.Printf("failed to get HTTP host address: %s", err)
		return
	}

	client, err := vaultClient.New(
		vaultClient.WithAddress(hostAddress),
		vaultClient.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Printf("failed to create Vault client: %s", err)
		return
	}

	if err := client.SetToken(token); err != nil {
		log.Printf("failed to set Vault token: %s", err)
		return
	}

	s, err := client.Secrets.KvV2Read(ctx, "test1", vaultClient.WithMountPath("secret"))
	if err != nil {
		log.Printf("failed to read secret: %s", err)
		return
	}

	fmt.Println(s.Data.Data["foo1"])
	// }

	// Output:
	// bar1
}
