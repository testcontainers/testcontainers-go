package elasticsearch_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"testing"

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

			opts := []testcontainers.ContainerCustomizer{testcontainers.WithImage(tt.image)}

			if tt.passwordCustomiser != nil {
				opts = append(opts, tt.passwordCustomiser)
			}

			esContainer, err := elasticsearch.RunContainer(ctx, opts...)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				if err := esContainer.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			httpClient := configureHTTPClient(esContainer)

			req, err := http.NewRequest("GET", esContainer.Settings.Address, nil)
			if err != nil {
				t.Fatal(err)
			}

			// set the password for the request using the Authentication header
			if tt.passwordCustomiser != nil {
				if esContainer.Settings.Username != "elastic" {
					t.Fatal("expected username to be elastic but got", esContainer.Settings.Username)
				}

				// basicAuthHeader {
				req.SetBasicAuth(esContainer.Settings.Username, esContainer.Settings.Password)
				// }
			}

			resp, err := httpClient.Do(req)
			if resp != nil {
				defer resp.Body.Close()
			}

			if tt.image != baseImage8 && err != nil {
				if tt.passwordCustomiser != nil {
					t.Fatal(err, "should access with authorised HTTP client.")
				} else if tt.passwordCustomiser == nil {
					t.Fatal(err, "should access with unauthorised HTTP client.")
				}
			}

			if tt.image == baseImage8 {
				if tt.passwordCustomiser != nil && err != nil {
					t.Fatal(err, "should access with authorised HTTP client using TLS.")
				}
				if tt.passwordCustomiser == nil && err == nil {
					// Elasticsearch 8 should return 401 Unauthorized, not an error in the request
					if resp.StatusCode != http.StatusUnauthorized {
						t.Fatal("expected 401 status code for unauthorised HTTP client using TLS, but got", resp.StatusCode)
					}

					// finish validating the response when the request is unauthorised
					return
				}

			}

			// validate response
			if resp != nil {
				// validate Elasticsearch response
				if resp.StatusCode != http.StatusOK {
					t.Fatal("expected 200 status code but got", resp.StatusCode)
				}

				var esResp ElasticsearchResponse
				if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
					t.Fatal(err)
				}

				if tt.image == baseImage7 && esResp.Version.Number != "7.9.2" {
					t.Fatal("expected version to be 7.9.2 but got", esResp.Version.Number)
				} else if tt.image == baseImage8 && esResp.Version.Number != "8.9.0" {
					t.Fatal("expected version to be 8.9.0 but got", esResp.Version.Number)
				}

				if esResp.Tagline != "You Know, for Search" {
					t.Fatal("expected tagline to be 'You Know, for Search' but got", esResp.Tagline)
				}
			}
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
			container, err := elasticsearch.RunContainer(
				ctx,
				testcontainers.WithImage(baseImage8),
				testcontainers.WithEnv(map[string]string{
					test.configKey: "false",
				}))
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				if err := container.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			if len(container.Settings.CACert) > 0 {
				t.Fatal("expected CA cert to be empty")
			}
		})
	}

}

func TestElasticsearch8WithoutCredentials(t *testing.T) {
	ctx := context.Background()

	container, err := elasticsearch.RunContainer(ctx, testcontainers.WithImage(baseImage8))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	httpClient := configureHTTPClient(container)

	req, err := http.NewRequest("GET", container.Settings.Address, nil)
	if err != nil {
		t.Fatal(err)
	}

	// elastic:changeme are the default credentials for Elasticsearch 8
	req.SetBasicAuth(container.Settings.Username, container.Settings.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err, "Should be able to access / URI with client using default password over HTTPS.")
	}

	defer resp.Body.Close()

	var esResp ElasticsearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		t.Fatal(err)
	}

	if esResp.Tagline != "You Know, for Search" {
		t.Fatal("expected tagline to be 'You Know, for Search' but got", esResp.Tagline)
	}
}

func TestElasticsearchOSSCannotuseWithPassword(t *testing.T) {
	ctx := context.Background()

	ossImage := elasticsearch.DefaultBaseImageOSS + ":7.9.2"

	_, err := elasticsearch.RunContainer(ctx, testcontainers.WithImage(ossImage), elasticsearch.WithPassword("foo"))
	if err == nil {
		t.Fatal(err, "Should not be able to use WithPassword with OSS image.")
	}
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
