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
