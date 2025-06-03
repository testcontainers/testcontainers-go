package elasticsearch_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

const (
	baseImage6 = "docker.elastic.co/elasticsearch/elasticsearch:6.8.23"
	baseImage7 = "docker.elastic.co/elasticsearch/elasticsearch:7.9.2"
	baseImage8 = "docker.elastic.co/elasticsearch/elasticsearch:8.9.0"
)

type ElasticsearchResponse struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number string `json:"number"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

func TestElasticsearch(t *testing.T) {
	// to be used in the container definition and in the HTTP client
	password := "foo"

	tests := []struct {
		name               string
		image              string
		passwordCustomiser testcontainers.ContainerCustomizer
	}{
		{
			name:               "Elasticsearch 6 without password should allow access using unauthenticated HTTP requests",
			image:              baseImage6,
			passwordCustomiser: nil,
		},
		{
			name:               "Elasticsearch 6 with password should allow access using authenticated HTTP requests",
			image:              baseImage6,
			passwordCustomiser: elasticsearch.WithPassword(password),
		},
		{
			name:               "Elasticsearch 7 without password should allow access using unauthenticated HTTP requests",
			image:              baseImage7,
			passwordCustomiser: nil,
		},
		{
			name:               "Elasticsearch 7 with password should allow access using authenticated HTTP requests",
			image:              baseImage7,
			passwordCustomiser: elasticsearch.WithPassword(password),
		},
		{
			name:               "Elasticsearch 8 without password should not allow access with unauthenticated HTTPS requests",
			image:              baseImage8,
			passwordCustomiser: nil,
		},
		{
			name:               "Elasticsearch 8 with password should allow access using authenticated HTTPS requests",
			image:              baseImage8,
			passwordCustomiser: elasticsearch.WithPassword(password),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			opts := []testcontainers.ContainerCustomizer{}

			if tt.passwordCustomiser != nil {
				opts = append(opts, tt.passwordCustomiser)
			}

			esContainer, err := elasticsearch.Run(ctx, tt.image, opts...)
			testcontainers.CleanupContainer(t, esContainer)
			require.NoError(t, err)

			httpClient := configureHTTPClient(esContainer)

			req, err := http.NewRequest(http.MethodGet, esContainer.Settings.Address, nil)
			require.NoError(t, err)

			// set the password for the request using the Authentication header
			if tt.passwordCustomiser != nil {
				require.Equalf(t, "elastic", esContainer.Settings.Username, "expected username to be elastic but got: %s", esContainer.Settings.Username)

				// basicAuthHeader {
				req.SetBasicAuth(esContainer.Settings.Username, esContainer.Settings.Password)
				// }
			}

			resp, err := httpClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			if tt.image == baseImage8 && tt.passwordCustomiser == nil {
				// Elasticsearch 8 should return 401 Unauthorized, not an error in the request
				require.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 status code for unauthorised HTTP client using TLS, but got: %s", resp.StatusCode)

				// finish validating the response when the request is unauthorised
				return
			}

			// validate Elasticsearch response
			require.Equalf(t, http.StatusOK, resp.StatusCode, "expected 200 status code but got: %s", resp.StatusCode)

			var esResp ElasticsearchResponse
			err = json.NewDecoder(resp.Body).Decode(&esResp)
			require.NoError(t, err)

			switch tt.image {
			case baseImage7:
				require.Equalf(t, "7.9.2", esResp.Version.Number, "expected version to be 7.9.2 but got: %s", esResp.Version.Number)
			case baseImage8:
				require.Equalf(t, "8.9.0", esResp.Version.Number, "expected version to be 8.9.0 but got: %s", esResp.Version.Number)
			}

			require.Equalf(t, "You Know, for Search", esResp.Tagline, "expected tagline to be 'You Know, for Search' but got: %s", esResp.Tagline)
		})
	}
}

func TestElasticsearch8WithoutSSL(t *testing.T) {
	tests := []struct {
		name      string
		configKey string
	}{
		{
			name:      "security disabled",
			configKey: "xpack.security.enabled",
		},
		{
			name:      "transport ssl disabled",
			configKey: "xpack.security.transport.ssl.enabled",
		},
		{
			name:      "http ssl disabled",
			configKey: "xpack.security.http.ssl.enabled",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ctr, err := elasticsearch.Run(
				ctx,
				baseImage8,
				testcontainers.WithEnv(map[string]string{
					test.configKey: "false",
				}))
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			require.Emptyf(t, ctr.Settings.CACert, "expected CA cert to be empty")
		})
	}
}

func TestElasticsearch8WithoutCredentials(t *testing.T) {
	ctx := context.Background()

	ctr, err := elasticsearch.Run(ctx, baseImage8)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpClient := configureHTTPClient(ctr)

	req, err := http.NewRequest(http.MethodGet, ctr.Settings.Address, nil)
	require.NoError(t, err)

	// elastic:changeme are the default credentials for Elasticsearch 8
	req.SetBasicAuth(ctr.Settings.Username, ctr.Settings.Password)

	resp, err := httpClient.Do(req)
	require.NoErrorf(t, err, "Should be able to access / URI with client using default password over HTTPS.")

	defer resp.Body.Close()

	var esResp ElasticsearchResponse
	err = json.NewDecoder(resp.Body).Decode(&esResp)
	require.NoError(t, err)

	require.Equalf(t, "You Know, for Search", esResp.Tagline, "expected tagline to be 'You Know, for Search' but got: %s", esResp.Tagline)
}

func TestElasticsearchOSSCannotUseWithPassword(t *testing.T) {
	ctx := context.Background()

	ossImage := elasticsearch.DefaultBaseImageOSS + ":7.9.2"

	ctr, err := elasticsearch.Run(ctx, ossImage, elasticsearch.WithPassword("foo"))
	testcontainers.CleanupContainer(t, ctr)
	require.Errorf(t, err, "Should not be able to use WithPassword with OSS image.")
}

// configureHTTPClient configures an HTTP client for the Elasticsearch container.
// If no certificate bytes are available, the default HTTP client will be returned.
// If certificate bytes are available, the client will be configured to use TLS with the certificate.
func configureHTTPClient(esContainer *elasticsearch.ElasticsearchContainer) *http.Client {
	// createHTTPClient {
	client := http.DefaultClient

	if esContainer.Settings.CACert == nil {
		return client
	}

	// configure TLS transport based on the certificate bytes that were retrieved from the container
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(esContainer.Settings.CACert)

	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	// }
	return client
}
