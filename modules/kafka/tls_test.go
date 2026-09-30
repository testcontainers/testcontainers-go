package kafka

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/testcontainers/testcontainers-go"
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

func TestEnsureTLSCerts(t *testing.T) {
	settings := &options{tlsEnabled: true}

	certs, err := settings.ensureTLSCerts("localhost")
	require.NoError(t, err)

	// a restart must reuse the certificates, so the keystore
	// in the container keeps matching TLSConfig()
	restartCerts, err := settings.ensureTLSCerts("localhost")
	require.NoError(t, err)
	require.Same(t, certs, restartCerts)
}

func TestConfigureSSLListener(t *testing.T) {
	tests := []struct {
		name                string
		env                 map[string]string
		expectedListeners   string
		expectedProtocolMap string
		wantErr             bool
	}{
		{
			name:                "no listeners",
			env:                 nil,
			expectedListeners:   "SSL://0.0.0.0:9095",
			expectedProtocolMap: "SSL:SSL",
		},
		{
			name: "listener appended",
			env: map[string]string{
				"KAFKA_LISTENERS":                      "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092",
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			},
			expectedListeners:   "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,SSL://0.0.0.0:9095",
			expectedProtocolMap: "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT,SSL:SSL",
		},
		{
			name: "listener already defined on the SSL port",
			env: map[string]string{
				"KAFKA_LISTENERS":                      "PLAINTEXT://0.0.0.0:9093, SSL://:9095",
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT, SSL:SSL",
			},
			expectedListeners:   "PLAINTEXT://0.0.0.0:9093, SSL://:9095",
			expectedProtocolMap: "PLAINTEXT:PLAINTEXT, SSL:SSL",
		},
		{
			name: "SASL_SSL is not SSL",
			env: map[string]string{
				"KAFKA_LISTENERS":                      "SASL_SSL://0.0.0.0:9096",
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "SASL_SSL:SASL_SSL",
			},
			expectedListeners:   "SASL_SSL://0.0.0.0:9096,SSL://0.0.0.0:9095",
			expectedProtocolMap: "SASL_SSL:SASL_SSL,SSL:SSL",
		},
		{
			name: "listener already defined on another port",
			env: map[string]string{
				"KAFKA_LISTENERS": "PLAINTEXT://0.0.0.0:9093,SSL://0.0.0.0:9096",
			},
			wantErr: true,
		},
		{
			name: "SSL listener mapped to another protocol",
			env: map[string]string{
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,SSL:PLAINTEXT",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env: test.env,
				},
			}

			err := configureSSLListener()(req)
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expectedListeners, req.Env["KAFKA_LISTENERS"])
			require.Equal(t, test.expectedProtocolMap, req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"])
		})
	}
}
