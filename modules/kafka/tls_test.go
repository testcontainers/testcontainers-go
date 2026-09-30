package kafka

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"
)

func TestCreateTLSCerts(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		expectedHosts []string
	}{
		{
			name:          "localhost",
			host:          "localhost",
			expectedHosts: []string{"localhost", "127.0.0.1"},
		},
		{
			name:          "remote docker host by name",
			host:          "docker.example.com",
			expectedHosts: []string{"localhost", "127.0.0.1", "docker.example.com"},
		},
		{
			name:          "remote docker host by IP",
			host:          "10.0.0.5",
			expectedHosts: []string{"localhost", "127.0.0.1", "10.0.0.5"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			certs, err := createTLSCerts(test.host)
			require.NoError(t, err)
			require.NotNil(t, certs.TLSConfig)

			key, cert, caCerts, err := pkcs12.DecodeChain(certs.KeystoreBytes, keystorePassword)
			require.NoError(t, err)
			require.NotNil(t, key)
			require.Len(t, caCerts, 1)

			for _, host := range test.expectedHosts {
				_, err := cert.Verify(x509.VerifyOptions{
					DNSName: host,
					Roots:   certs.TLSConfig.RootCAs,
				})
				require.NoErrorf(t, err, "expected certificate to be valid for %s", host)
			}
		})
	}
}

func TestAppendListEntry(t *testing.T) {
	tests := []struct {
		name     string
		list     string
		prefix   string
		entry    string
		expected string
	}{
		{
			name:     "empty list",
			list:     "",
			prefix:   "SSL://",
			entry:    "SSL://0.0.0.0:9095",
			expected: "SSL://0.0.0.0:9095",
		},
		{
			name:     "listener appended",
			list:     "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092",
			prefix:   "SSL://",
			entry:    "SSL://0.0.0.0:9095",
			expected: "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,SSL://0.0.0.0:9095",
		},
		{
			name:     "listener already defined",
			list:     "PLAINTEXT://0.0.0.0:9093, SSL://0.0.0.0:9096",
			prefix:   "SSL://",
			entry:    "SSL://0.0.0.0:9095",
			expected: "PLAINTEXT://0.0.0.0:9093, SSL://0.0.0.0:9096",
		},
		{
			name:     "SASL_SSL is not SSL",
			list:     "BROKER:PLAINTEXT,SASL_SSL:SASL_SSL",
			prefix:   "SSL:",
			entry:    "SSL:SSL",
			expected: "BROKER:PLAINTEXT,SASL_SSL:SASL_SSL,SSL:SSL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, appendListEntry(test.list, test.prefix, test.entry))
		})
	}
}
