package vault

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	vaultClient "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

const (
	token = "root-token"
)

var (
	ctx   = context.Background()
	vault *vaultContainer
)

func TestMain(m *testing.M) {
	var err error
	// StartContainer {
	vault, err = StartContainer(ctx,
		WithLogLevel(Debug),
		WithToken(token),
		WithInitCommand("secrets enable transit", "write -f transit/keys/my-key"),
		WithInitCommand("kv put secret/test1 foo1=bar1"),
		WithInitCommand("kv put secret/test2 foo2=bar2 foo3=bar3"))
	// }
	if err != nil {
		log.Fatal(err)
	}

	c := m.Run()

	// Clean up the vault after the test is complete
	if err = vault.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate vault: %s", err)
	}

	os.Exit(c)
}

func TestVaultFirstSecretPathWithCLI(t *testing.T) {
	exec, reader, err := vault.Exec(ctx, []string{"vault", "kv", "get", "-format=json", "secret/test1"})
	assert.Nil(t, err)
	assert.Equal(t, 0, exec)

	bytes, err := io.ReadAll(reader)
	assert.Nil(t, err)

	assert.Equal(t, "bar1", gjson.Get(string(bytes), "data.data.foo1").String())
}

func TestVaultSecondSecretPathWithCLI(t *testing.T) {
	exec, reader, err := vault.Exec(ctx, []string{"vault", "kv", "get", "-format=json", "secret/test2"})
	assert.Nil(t, err)
	assert.Equal(t, 0, exec)

	bytes, err := io.ReadAll(reader)
	assert.Nil(t, err)

	assert.Equal(t, "bar2", gjson.Get(string(bytes), "data.data.foo2").String())
	assert.Equal(t, "bar3", gjson.Get(string(bytes), "data.data.foo3").String())
}

func TestVaultFirstSecretPathWithHTTP(t *testing.T) {
	hostAddress, err := vault.HttpHostAddress(ctx)
	assert.Nil(t, err)

	request, _ := http.NewRequest(http.MethodGet, hostAddress+"/v1/secret/data/test1", nil)
	request.Header.Add("X-Vault-Token", token)

	response, err := http.DefaultClient.Do(request)
	assert.Nil(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	assert.Nil(t, err)

	assert.Equal(t, "bar1", gjson.Get(string(body), "data.data.foo1").String())
}

func TestVaultSecondSecretPathWithHTTP(t *testing.T) {
	hostAddress, err := vault.HttpHostAddress(ctx)
	assert.Nil(t, err)

	request, _ := http.NewRequest(http.MethodGet, hostAddress+"/v1/secret/data/test2", nil)
	request.Header.Add("X-Vault-Token", token)

	response, err := http.DefaultClient.Do(request)
	assert.Nil(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	assert.Nil(t, err)

	assert.Equal(t, "bar2", gjson.Get(string(body), "data.data.foo2").String())
	assert.Equal(t, "bar3", gjson.Get(string(body), "data.data.foo3").String())
}

func TestVaultFirstSecretPathWithClient(t *testing.T) {
	hostAddress, _ := vault.HttpHostAddress(ctx)
	client, err := vaultClient.New(
		vaultClient.WithAddress(hostAddress),
		vaultClient.WithRequestTimeout(30*time.Second),
	)
	assert.Nil(t, err)

	err = client.SetToken(token)
	assert.Nil(t, err)

	s, err := client.Secrets.KVv2Read(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "bar1", s.Data["data"].(map[string]interface{})["foo1"])
}

func TestVaultSecondSecretPathWithClient(t *testing.T) {
	hostAddress, _ := vault.HttpHostAddress(ctx)
	client, err := vaultClient.New(
		vaultClient.WithAddress(hostAddress),
		vaultClient.WithRequestTimeout(30*time.Second),
	)
	assert.Nil(t, err)

	err = client.SetToken(token)
	assert.Nil(t, err)

	s, err := client.Secrets.KVv2Read(ctx, "test2")
	assert.Nil(t, err)
	assert.Equal(t, "bar2", s.Data["data"].(map[string]interface{})["foo2"])
	assert.Equal(t, "bar3", s.Data["data"].(map[string]interface{})["foo3"])
}

func TestVaultWriteSecretWithClient(t *testing.T) {
	hostAddress, _ := vault.HttpHostAddress(ctx)
	client, err := vaultClient.New(
		vaultClient.WithAddress(hostAddress),
		vaultClient.WithRequestTimeout(30*time.Second),
	)
	assert.Nil(t, err)

	err = client.SetToken(token)
	assert.Nil(t, err)

	_, err = client.Secrets.KVv2Write(ctx, "test3", schema.KVv2WriteRequest{
		Data: map[string]any{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)

	s, err := client.Secrets.KVv2Read(ctx, "test3")
	assert.Nil(t, err)
	assert.Equal(t, "bar", s.Data["data"].(map[string]interface{})["foo"])
}
