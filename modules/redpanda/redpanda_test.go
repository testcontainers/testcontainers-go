package redpanda

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestRedpanda(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	require.NoError(t, err)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Test Kafka API
	seedBroker, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(seedBroker),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	kafkaAdmCl := kadm.NewClient(kafkaCl)
	metadata, err := kafkaAdmCl.Metadata(ctx)
	require.NoError(t, err)
	assert.Len(t, metadata.Brokers, 1)

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Admin API
	// adminAPIAddress {
	adminAPIURL, err := container.AdminAPIAddress(ctx)
	// }
	require.NoError(t, err)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/cluster/health_overview", adminAPIURL), nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test produce to unknown topic
	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.Error(t, results.FirstErr(), kerr.UnknownTopicOrPartition)
}

func TestRedpandaWithAuthentication(t *testing.T) {
	ctx := context.Background()
	// redpandaCreateContainer {
	container, err := RunContainer(ctx,
		WithEnableSASL(),
		WithEnableKafkaAuthorization(),
		WithNewServiceAccount("superuser-1", "test"),
		WithNewServiceAccount("superuser-2", "test"),
		WithNewServiceAccount("no-superuser", "test"),
		WithSuperusers("superuser-1", "superuser-2"),
		WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	require.NoError(t, err)
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// kafkaSeedBroker {
	seedBroker, err := container.KafkaSeedBroker(ctx)
	// }
	require.NoError(t, err)

	// Test successful authentication & authorization with all created superusers
	serviceAccounts := map[string]string{
		"superuser-1": "test",
		"superuser-2": "test",
	}

	for user, password := range serviceAccounts {
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: user,
				Pass: password,
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, fmt.Sprintf("test-%v", user))
		require.NoError(t, err)
		kafkaCl.Close()
	}

	// Test successful authentication, but failed authorization with a non-superuser account
	{
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "no-superuser",
				Pass: "test",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.CreateTopic(ctx, 1, 1, nil, "test-2")
		require.Error(t, err)
		require.ErrorContains(t, err, "TOPIC_AUTHORIZATION_FAILED")
		kafkaCl.Close()
	}

	// Test failed authentication
	{
		kafkaCl, err := kgo.NewClient(
			kgo.SeedBrokers(seedBroker),
			kgo.SASL(scram.Auth{
				User: "wrong",
				Pass: "wrong",
			}.AsSha256Mechanism()),
		)
		require.NoError(t, err)

		kafkaAdmCl := kadm.NewClient(kafkaCl)
		_, err = kafkaAdmCl.Metadata(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "SASL_AUTHENTICATION_FAILED")
	}

	// Test Schema Registry API
	httpCl := &http.Client{Timeout: 5 * time.Second}
	// schemaRegistryAddress {
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	// }
	require.NoError(t, err)

	// Failed authentication
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Successful authentication
	for user, password := range serviceAccounts {
		req.SetBasicAuth(user, password)
		resp, err = httpCl.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRedpandaProduceWithAutoCreateTopics(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, WithAutoCreateTopics())
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	brokers, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.AllowAutoTopicCreation(),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.NoError(t, results.FirstErr())
}

func TestRedpandaWithTLS(t *testing.T) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	require.NoError(t, err, "failed to load key pair")

	ctx := context.Background()

	container, err := RunContainer(ctx, WithTLS(localhostCert, localhostKey))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(localhostCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	httpCl := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
		},
	}

	// Test Admin API
	adminAPIURL, err := container.AdminAPIAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "AdminAPIAddress should return https url")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/cluster/health_overview", adminAPIURL), nil)
	require.NoError(t, err)
	resp, err := httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test Schema Registry API
	schemaRegistryURL, err := container.SchemaRegistryAddress(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(adminAPIURL, "https://"), "SchemaRegistryAddress should return https url")
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/subjects", schemaRegistryURL), nil)
	require.NoError(t, err)
	resp, err = httpCl.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	brokers, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	kafkaCl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.DialTLSConfig(tlsConfig),
	)
	require.NoError(t, err)
	defer kafkaCl.Close()

	// Test produce to unknown topic
	results := kafkaCl.ProduceSync(ctx, &kgo.Record{Topic: "test", Value: []byte("test message")})
	require.Error(t, results.FirstErr(), kerr.UnknownTopicOrPartition)
}

func TestRedpandaListener_Simple(t *testing.T) {
	ctx := context.Background()

	// 1. Create network
	rpNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	require.NoError(t, err)

	// 2. Start Redpanda container
	// withListenerRP {
	container, err := RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		network.WithNetwork([]string{"redpanda-host"}, rpNetwork),
		WithListener("redpanda:29092"), WithAutoCreateTopics(),
	)
	// }
	require.NoError(t, err)

	// 3. Start KCat container
	// withListenerKcat {
	kcat, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "confluentinc/cp-kcat:7.4.1",
			Networks: []string{
				rpNetwork.Name,
			},
			Entrypoint: []string{
				"sh",
			},
			Cmd: []string{
				"-c",
				"tail -f /dev/null",
			},
		},
		Started: true,
	})
	// }

	require.NoError(t, err)

	// 4. Copy message to kcat
	err = kcat.CopyToContainer(ctx, []byte("Message produced by kcat"), "/tmp/msgs.txt", 700)
	require.NoError(t, err)

	// 5. Produce message to Redpanda
	// withListenerExec {
	_, _, err = kcat.Exec(ctx, []string{"kcat", "-b", "redpanda:29092", "-t", "msgs", "-P", "-l", "/tmp/msgs.txt"})
	// }

	require.NoError(t, err)

	// 6. Consume message from Redpanda
	_, stdout, err := kcat.Exec(ctx, []string{"kcat", "-b", "redpanda:29092", "-C", "-t", "msgs", "-c", "1"})
	require.NoError(t, err)

	// 7. Read Message from stdout
	out, err := io.ReadAll(stdout)
	require.NoError(t, err)

	require.Contains(t, string(out), "Message produced by kcat")

	t.Cleanup(func() {
		if err := kcat.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate kcat container: %s", err)
		}
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate redpanda container: %s", err)
		}

		if err := rpNetwork.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %s", err)
		}
	})
}

func TestRedpandaListener_InvalidPort(t *testing.T) {
	ctx := context.Background()

	// 1. Create network
	RPNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	require.NoError(t, err)

	// 2. Attempt Start Redpanda container
	_, err = RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		WithListener("redpanda:99092"),
		network.WithNetwork([]string{"redpanda-host"}, RPNetwork),
	)

	require.Error(t, err)

	require.Contains(t, err.Error(), "invalid port on listener redpanda:99092")

	t.Cleanup(func() {
		if err := RPNetwork.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %s", err)
		}
	})
}

func TestRedpandaListener_NoNetwork(t *testing.T) {
	ctx := context.Background()

	// 1. Attempt Start Redpanda container
	_, err := RunContainer(ctx,
		testcontainers.WithImage("redpandadata/redpanda:v23.2.18"),
		WithListener("redpanda:99092"),
	)

	require.Error(t, err)

	require.Contains(t, err.Error(), "container must be attached to at least one network")
}

// localhostCert is a PEM-encoded TLS cert with SAN IPs
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 2048 --host 127.0.0.1,::1,localhost --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDODCCAiCgAwIBAgIRAKMykg5qJSCb4L3WtcZznSQwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPYcLIhqCsrmqvsY1gWqI1jx3Ytn5Qjfvlg3BPD/YeD4UVouBhgQ
NIIERFCmDUzu52pXYZeCouBIVDWqZKixQf3PyBzAqbFvX0pTsZrOnvjuoahzjEcl
x+CfkIp58mVaV/8v9TyBYCXNuHlI7Pndu/3U5d6npSg8+dTkwW3VZzZyHpsDW+a4
ByW02NI58LoHzQPMRg9MFToL1qNQy4PFyADf2N/3/SYOkrbSrXA0jYqXE8yvQGYe
LWcoQ+4YkurSS1TgSNEKxrzGj8w4xRjEjRNsLVNWd8uxZkHwv6LXOn4s39ix3jN4
7OJJHA8fJAWxAP4ThrpM1j5J+Rq1PD380u8CAwEAAaOBhjCBgzAOBgNVHQ8BAf8E
BAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zAdBgNV
HQ4EFgQU8gMBt2leRAnGgCQ6pgIYPHY35GAwLAYDVR0RBCUwI4IJbG9jYWxob3N0
hwR/AAABhxAAAAAAAAAAAAAAAAAAAAABMA0GCSqGSIb3DQEBCwUAA4IBAQA5F6aw
6JJMsnCjxRGYXb252zqjxOxweawZ2je4UAGSsF27Phm1Bx6/2mzPpgIB0I7xNBFL
ljtqBG/FpH6qWpkkegljL8Z5soXiye/4r1G+V6hadm32/OLQCS//dyq7W1a2uVlS
KdFjoNqRW2PacVQLjnTbP2SJV5CnrJgCsSMXVoNnKdj5gr5ltNNAt9TAJ85iFa5d
rJla/XghtqEOzYtigKPF7EVqRRl4RmPu30hxwDZMT60ptFolfCEeXpDra5uonJMv
ElEbzK8ZzXmvWCj94RjPkGKZs8+SDM2qfKPk5ZW2xJxwqS3tkEkZlj1L+b7zYOlt
aJ65OWCXHLecrgdl
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(testingKey(`-----BEGIN TESTING KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQD2HCyIagrK5qr7
GNYFqiNY8d2LZ+UI375YNwTw/2Hg+FFaLgYYEDSCBERQpg1M7udqV2GXgqLgSFQ1
qmSosUH9z8gcwKmxb19KU7Gazp747qGoc4xHJcfgn5CKefJlWlf/L/U8gWAlzbh5
SOz53bv91OXep6UoPPnU5MFt1Wc2ch6bA1vmuAcltNjSOfC6B80DzEYPTBU6C9aj
UMuDxcgA39jf9/0mDpK20q1wNI2KlxPMr0BmHi1nKEPuGJLq0ktU4EjRCsa8xo/M
OMUYxI0TbC1TVnfLsWZB8L+i1zp+LN/Ysd4zeOziSRwPHyQFsQD+E4a6TNY+Sfka
tTw9/NLvAgMBAAECggEBALKxAiSJ2gw4Lyzhe4PhZIjQE+uEI+etjKbAS/YvdwHB
SlAP2pzeJ0G/l1p3NnEFhUDQ8SrwzxHJclsEvNE+4otGsiUuPgd2tdlhqzKbkxFr
MjT8sH14EQgm0uu4Xyb30ayXRZgI16abF7X4HRfOxxAl5EElt+TfYQYSkd8Nc0Mz
bD7g0riSdOKVhNIkUTT1U7x8ClIgff6vbWztOVP4hGezqEKpO/8+JBkg2GLeH3lC
PyuHEb33Foxg7SX35M1a89EKC2p4ER6/nfg6wGYyIsn42gBk1JgQdg23x7c/0WOu
vcw1unNP2kCbnsCeZ6KPRRGXEjbpTqOTzAUOekOeOgECgYEA9/jwK2wrX2J3kJN7
v6kmxazigXHCa7XmFMgTfdqiQfXfjdi/4u+4KAX04jWra3ZH4KT98ztPOGjb6KhM
hfMldsxON8S9IQPcbDyj+5R77KU4BG/JQBEOX1uzS9KjMVG5e9ZUpG5UnSoSOgyM
oN3DZto7C5ULO2U2MT8JaoGb53cCgYEA/hPNMsCXFairxKy0BCsvJFan93+GIdwM
YoAGLc4Oj67ES8TYC4h9Im5i81JYOjpY4aZeKdj8S+ozmbqqa/iJiAfOr37xOMuX
AQA2T8uhPXXNXA5s6T3LaIXtzL0NmRRZCtuyEGdCidIXub7Bz8LrfsMc+s/jv57f
4IPmW12PPkkCgYBpEdDqBT5nfzh8SRGhR1IHZlbfVE12CDACVDh2FkK0QjNETjgY
N0zHoKZ/hxAoS4jvNdnoyxOpKj0r2sv54enY6X6nALTGnXUzY4p0GhlcTzFqJ9eV
TuTRIPDaytidGCzIvStGNP2jTmVEtXaM3wphtUxZfwCwXRVWToh12Y8uxwKBgA1a
FQp5vHbS6lPnj344lr2eIC2NcgsNeUkj2S9HCNTcJkylB4Vzor/QdTq8NQ66Sjlx
eLlSQc/retK1UIdkBDY10tK+JQcLC+Btlm0TEmIccrJHv8lyCeJwR1LfDHvi6dr8
OJtMEd8UP1Lvh1fXsnBy6G71xc4oFzPBOrXKcOChAoGACOgyYe47ZizScsUGjCC7
xARTEolZhtqHKVd5s9oi95P0r7A1gcNx/9YW0rCT2ZD8BD9H++HTE2L+mh3R9zDn
jwDeW7wVZec+oyGdc9L+B1xU25O+88zNLxlRAX8nXJbHdgL83UclmC51GbXejloP
D4ZNvyXf/6E27Ibu6v2p/vs=
-----END TESTING KEY-----`))

func testingKey(s string) string { return strings.ReplaceAll(s, "TESTING KEY", "PRIVATE KEY") }
